// Package reservations lets guests reserve or mark purchase of gift items
// without an account, safely under concurrency.
package reservations

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/listou/listou/apps/api/internal/platform/validate"
)

type Reservation struct {
	ID           uuid.UUID  `json:"id"`
	ItemID       uuid.UUID  `json:"itemId"`
	Kind         string     `json:"kind"`
	Quantity     int        `json:"quantity"`
	Status       string     `json:"status"`
	GuestName    string     `json:"-"`
	GuestContact *string    `json:"-"`
	Message      *string    `json:"-"`
	ExpiresAt    *time.Time `json:"expiresAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	// ManageToken is only set on the creation response.
	ManageToken string `json:"manageToken,omitempty"`
}

type CreateInput struct {
	Kind         string  `json:"kind"`
	Quantity     int     `json:"quantity"`
	GuestName    string  `json:"guestName"`
	GuestContact *string `json:"guestContact"`
	Message      *string `json:"message"`
}

func (in *CreateInput) Validate() error {
	errs := validate.Errors{}
	in.GuestName = strings.TrimSpace(in.GuestName)
	in.GuestContact, in.Message = validate.Trim(in.GuestContact), validate.Trim(in.Message)
	if in.Kind == "" {
		in.Kind = "RESERVATION"
	}
	if in.Kind != "RESERVATION" && in.Kind != "PURCHASE" {
		errs.Add("kind", "Tipo inválido.")
	}
	if in.Quantity == 0 {
		in.Quantity = 1
	}
	if in.Quantity < 1 || in.Quantity > 99 {
		errs.Add("quantity", "Quantidade entre 1 e 99.")
	}
	errs.Length("guestName", in.GuestName, 1, 80, "Como devemos te chamar?")
	if in.GuestContact != nil {
		errs.Length("guestContact", *in.GuestContact, 0, 120, "Contato muito longo.")
	}
	if in.Message != nil {
		errs.Length("message", *in.Message, 0, 500, "Mensagem muito longa.")
	}
	return errs.Err()
}
