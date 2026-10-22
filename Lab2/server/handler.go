package main

import (
	"Lab2/server/storage"
	"Lab2/shared/protocol"
)

type Handler struct {
	store     *storage.KeyStore
	adminPass string
}

func NewHandler(store *storage.KeyStore, adminPass string) *Handler {
	return &Handler{
		store:     store,
		adminPass: adminPass,
	}
}

func (handler *Handler) Handle(raw string) string {
	request, err := protocol.ParseRequest(raw)
	if err != nil {
		return (&protocol.Response{
			NrID:     "",
			Contents: protocol.ContentError,
		}).String()
	}

	switch request.Command {
	case protocol.CmdGet:
		if handler.store.Contains(request.NrID) {
			return (&protocol.Response{
				NrID:     request.NrID,
				Contents: protocol.ContentYes,
			}).String()
		}
		return (&protocol.Response{
			NrID:     request.NrID,
			Contents: protocol.ContentNot,
		}).String()

	case protocol.CmdSet:
		if request.Password == handler.adminPass {
			handler.store.Add(request.NrID)
			return (&protocol.Response{
				NrID:     request.NrID,
				Contents: protocol.ContentYes,
			}).String()
		}
		return (&protocol.Response{
			NrID:     request.NrID,
			Contents: protocol.ContentNot,
		}).String()

	default:
		return (&protocol.Response{
			NrID:     "",
			Contents: protocol.ContentError,
		}).String()
	}
}
