package main

import (
	"bytes"
	"fmt"
	"io"
)

var participantId = 0

type Participant struct {
	Id     int
	Name   string
	Gender string
}

func newParticipant(name string, gender string) Participant {
	participantId++
	return Participant{
		Id:     participantId,
		Name:   name,
		Gender: gender,
	}
}

type Participants []Participant

func (p *Participants) getParticipantById(id int) *Participant {
	for _, p := range *p {
		if p.Id == id {
			return &p
		}
	}
	return nil
}

func (p *Participants) getGenders() []string {
	s := map[string]bool{}

	for _, part := range *p {
		s[part.Gender] = true
	}

	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}

	return keys
}

func newParticipants() Participants {
	return Participants{
		newParticipant("Audun", "m"),
		newParticipant("Mali", "f"),
	}
}

var talkTypeId = 0

type TalkType struct {
	Id         int
	Name       string
	MaxReplies int
	Color      string
}

func newTalkType(name string, replies int, color string) TalkType {
	talkTypeId++
	return TalkType{
		Id:         talkTypeId,
		Name:       name,
		MaxReplies: replies,
		Color:      color,
	}
}

type TalkTypes []TalkType

func (t *TalkTypes) getTalkTypeById(id int) *TalkType {
	for _, t := range *t {
		if t.Id == id {
			return &t
		}
	}
	return nil
}

func newTalkTypes() TalkTypes {
	return TalkTypes{
		newTalkType("Innlegg", 2, "#00FF00"),
		newTalkType("Replikk", 0, "#440C1A"),
		newTalkType("Til dagsorden", 0, "#0000FF"),
	}
}

var talkId = 0

type Talk struct {
	Id          int
	Type        int
	Participant int
	ReplyTo     *Talk
}

func newTalk(participantId int, talkTypeId int, replyTo *Talk) Talk {
	talkId++
	return Talk{
		Id:          talkId,
		Type:        talkTypeId,
		Participant: participantId,
		ReplyTo:     replyTo,
	}
}

type Talks []Talk

func newTalks() Talks {
	return Talks{}
}

type Page struct {
	TalkTypes    TalkTypes
	Participants Participants
	TalkBlocks   TalkBlocks
}

func newPage() Page {
	return Page{
		TalkTypes:    newTalkTypes(),
		Participants: newParticipants(),
		TalkBlocks:   newTalkBlocks(),
	}
}

type Message struct {
	message string
}

func newMessage(message string) Message {
	return Message{
		message: message,
	}
}

type TalkBlock struct {
	Talk        Talk
	Participant Participant
	TalkType    TalkType
}

type TalkBlocks []TalkBlock

func newTalkBlocks() TalkBlocks {
	return TalkBlocks{}
}

func newTalkBlock(talk Talk, participant Participant, talk_type TalkType) TalkBlock {
	return TalkBlock{
		Talk:        talk,
		Participant: participant,
		TalkType:    talk_type,
	}
}

// Event represents Server-Sent Event.
// SSE explanation: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
type Event struct {
	// ID is used to set the EventSource object's last event ID value.
	ID []byte
	// Data field is for the message. When the EventSource receives multiple consecutive lines
	// that begin with data:, it concatenates them, inserting a newline character between each one.
	// Trailing newlines are removed.
	Data []byte
	// Event is a string identifying the type of event described. If this is specified, an event
	// will be dispatched on the browser to the listener for the specified event name; the website
	// source code should use addEventListener() to listen for named events. The onmessage handler
	// is called if no event name is specified for a message.
	Event []byte
	// Retry is the reconnection time. If the connection to the server is lost, the browser will
	// wait for the specified time before attempting to reconnect. This must be an integer, specifying
	// the reconnection time in milliseconds. If a non-integer value is specified, the field is ignored.
	Retry []byte
	// Comment line can be used to prevent connections from timing out; a server can send a comment
	// periodically to keep the connection alive.
	Comment []byte
}

// MarshalTo marshals Event to given Writer
func (ev *Event) MarshalTo(w io.Writer) error {
	// Marshalling part is taken from: https://github.com/r3labs/sse/blob/c6d5381ee3ca63828b321c16baa008fd6c0b4564/http.go#L16
	if len(ev.Data) == 0 && len(ev.Comment) == 0 {
		return nil
	}

	if len(ev.Data) > 0 {
		if _, err := fmt.Fprintf(w, "id: %s\n", ev.ID); err != nil {
			return err
		}

		sd := bytes.Split(ev.Data, []byte("\n"))
		for i := range sd {
			if _, err := fmt.Fprintf(w, "data: %s\n", sd[i]); err != nil {
				return err
			}
		}

		if len(ev.Event) > 0 {
			if _, err := fmt.Fprintf(w, "event: %s\n", ev.Event); err != nil {
				return err
			}
		}

		if len(ev.Retry) > 0 {
			if _, err := fmt.Fprintf(w, "retry: %s\n", ev.Retry); err != nil {
				return err
			}
		}
	}

	if len(ev.Comment) > 0 {
		if _, err := fmt.Fprintf(w, ": %s\n", ev.Comment); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}

	return nil
}
