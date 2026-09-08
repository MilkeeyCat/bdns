package message

import "github.com/MilkeeyCat/bdns/record"

type Message struct {
	Header     Header
	Question   []Question
	Answer     []record.Record
	Authority  []record.Record
	Additional []record.Record
}
