package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"

	"golang.org/x/sync/errgroup"

	"github.com/MilkeeyCat/bdns/domain"
	"github.com/MilkeeyCat/bdns/record"
	"github.com/MilkeeyCat/bdns/wire"
)

type Server struct {
	port uint16
}

func New(port uint16) *Server {
	return &Server{
		port: port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	slog.Info("starting DNS server")

	g.Go(func() error {
		return s.listenUDP(ctx)
	})
	g.Go(func() error {
		return s.listenTCP(ctx)
	})

	return g.Wait()
}

func (s *Server) listenUDP(ctx context.Context) error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: int(s.port)})
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	for {
		var buf [wire.MaxUDPMessageSize]byte
		n, addr, err := conn.ReadFrom(buf[:])

		if n > 0 {
			go func() {
				res, err := s.processMessage(buf[:n])
				if err != nil {
					slog.Info(
						"failed to process message",
						slog.Any("error", err),
					)

					return
				}

				if _, err := conn.WriteTo(res, addr); err != nil {
					slog.Error(
						"failed to write UDP response",
						slog.Any("error", err),
					)
				}
			}()
		}

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			slog.Error("failed to read UDP packet", slog.Any("error", err))
		}
	}
}

func (s *Server) listenTCP(ctx context.Context) error {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{Port: int(s.port)})
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			slog.Error(
				"failed to accept TCP connection",
				slog.Any("error", err),
			)

			continue
		}

		go func() {
			defer func() { _ = conn.Close() }()

			var sizeBuf [2]byte

			if _, err := io.ReadFull(conn, sizeBuf[:]); err != nil {
				slog.Info(
					"failed to read message length",
					slog.Any("error", err),
				)

				return
			}

			size := binary.BigEndian.Uint16(sizeBuf[:])
			buf := make([]byte, size)

			if _, err := io.ReadFull(conn, buf); err != nil {
				slog.Info(
					"failed to read message contents",
					slog.Any("error", err),
				)

				return
			}

			res, err := s.processMessage(buf)
			if err != nil {
				slog.Info("failed to process message", slog.Any("error", err))

				return
			}

			binary.BigEndian.PutUint16(sizeBuf[:], uint16(len(res)))

			bufs := net.Buffers{sizeBuf[:], res}

			if _, err := bufs.WriteTo(conn); err != nil {
				slog.Error(
					"failed to write TCP response",
					slog.Any("error", err),
				)
			}
		}()
	}
}

func (s *Server) processMessage(buf []byte) ([]byte, error) {
	p := wire.NewParser(buf)
	msg, err := p.ParseMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	// dummy response
	{
		msg.Header.QR = true
		msg.Header.RA = false
		msg.Header.AA = true
		msg.Answer = []record.Record{
			{
				Name:  domain.NewDomainFromString("dummy.local."),
				Type:  record.TypeA,
				Class: record.ClassIN,
				TTL:   300,
				Data:  []byte{1, 2, 3, 4},
			},
		}
	}

	e := wire.NewEncoder(wire.WithMessageLimit(wire.MaxUDPMessageSize))

	if err := e.EncodeMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to encode message: %w", err)
	}

	return e.Bytes(), nil
}
