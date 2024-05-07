package main

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

func newParticipants() Participants {
	return Participants{
		newParticipant("Audun", "m"),
		newParticipant("Mali", "f"),
	}
}

var talkTypeId = 0

type TalkType struct {
	Id      int
	Name    string
	Replies bool
}

func newTalkType(name string, replies bool) TalkType {
	talkTypeId++
	return TalkType{
		Id:      talkTypeId,
		Name:    name,
		Replies: replies,
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
		newTalkType("Innlegg", true),
		newTalkType("Replikk", false),
		newTalkType("Til dagsorden", false),
	}
}

var talkId = 0

type Talk struct {
	Id          int
	Type        int
	Participant int
}

func newTalk(participantId int, talkTypeId int) Talk {
	talkId++
	return Talk{
		Id:          talkId,
		Type:        talkTypeId,
		Participant: participantId,
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
