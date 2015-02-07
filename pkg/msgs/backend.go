package msgs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"pkg/util"
)

//Backend - poruka koja dolazi iz backend servisa
type Backend struct {
	Type    string
	Id      string
	IgracId string
	No      int64
	From    string
	To      string
	IsDel   bool
	Gzip    bool   //da li je body inicijalno bio gzip-an
	Body    []byte //raspakovan body
}

func NewBackend(buf []byte) (*Backend, error) {
	return parseAsBackend(buf)
}

func CreateBackend(typ string, no int, body []byte) []byte {
	header := map[string]interface{}{
		"type": typ,
		"no":   no,
	}
	if len(body) > 1024 {
		body = util.Gzip(body)
		header["encoding"] = "gzip"
	}
	buf, _ := json.Marshal(header)
	buf = append(buf, []byte{10}...)
	buf = append(buf, body...)
	return buf
}

func parseAsBackend(buf []byte) (*Backend, error) {
	parts := bytes.SplitN(buf, []byte{10}, 2)
	rawHeader := parts[0]

	header := struct {
		DocType   string `json:"doc_type"`
		Type      string `json:"type"`
		DocId     string `json:"doc_id"`
		Id        string `json:"id"`
		DocAction string `json:"doc_action"`
		Action    string `json:"action"`
		IgracId   string `json:"igrac_id"`
		From      string `json:"from"`
		To        string `json:"to"`
		No        int64  `json:"no"`
		Encoding  string `json:"encoding"`
	}{
		No:      -1,
		IgracId: "*",
	}

	err := json.Unmarshal(rawHeader, &header)
	if err != nil {
		return nil, fmt.Errorf("error parsing json header: %s", rawHeader)
	}
	if header.Type == "" && header.DocType != "" {
		header.Type = header.DocType
	}
	if header.Id == "" && header.DocId != "" {
		header.Id = header.DocId
	}

	msg := &Backend{
		Type:    header.Type,
		IgracId: header.IgracId,
		Id:      header.Id,
		No:      header.No,
		IsDel:   header.DocAction == "del" || header.Action == "del",
		From:    header.From,
		To:      header.To,
		Gzip:    header.Encoding == "gzip",
	}
	if len(parts) > 1 {
		body := parts[1]
		if msg.Gzip && util.IsGziped(body) {
			if msg.Body, err = util.Gunzip(body); err != nil {
				return nil, err
			}
		} else {
			msg.Body = body
		}
	}
	return msg, nil
}

func (b *Backend) bodyStr() string {
	return string(b.Body)
}
