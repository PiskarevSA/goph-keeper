package gophkeeper

import (
	gophkeeperv1 "github.com/PiskarevSA/goph-keeper/gen/gophkeeper/v1"
	"github.com/PiskarevSA/goph-keeper/internal/domain"
)

// map domain.SecretKind to proto SecretInfo_Kind
func domainToPbInfoKind(k domain.SecretKind) gophkeeperv1.SecretInfo_Kind {
	switch k {
	case domain.KindCredentials:
		return gophkeeperv1.SecretInfo_KIND_CREDENTIALS
	case domain.KindCard:
		return gophkeeperv1.SecretInfo_KIND_CARD
	case domain.KindText:
		return gophkeeperv1.SecretInfo_KIND_TEXT
	case domain.KindRaw:
		return gophkeeperv1.SecretInfo_KIND_RAW
	default:
		return gophkeeperv1.SecretInfo_KIND_UNSPECIFIED
	}
}

// map proto SecretInfo_Kind to domain.SecretKind
func pbInfoKindToDomain(k gophkeeperv1.SecretInfo_Kind) domain.SecretKind {
	switch k {
	case gophkeeperv1.SecretInfo_KIND_CREDENTIALS:
		return domain.KindCredentials
	case gophkeeperv1.SecretInfo_KIND_CARD:
		return domain.KindCard
	case gophkeeperv1.SecretInfo_KIND_TEXT:
		return domain.KindText
	case gophkeeperv1.SecretInfo_KIND_RAW:
		return domain.KindRaw
	default:
		return domain.KindUnspecified
	}
}

// map proto SecretDetails.Secret.(type) to domain.SecretKind
func deriveKindFromPbDetails(d *gophkeeperv1.SecretDetails) domain.SecretKind {
	if d == nil {
		return domain.KindUnspecified
	}
	switch d.Secret.(type) {
	case *gophkeeperv1.SecretDetails_Credentials:
		return domain.KindCredentials
	case *gophkeeperv1.SecretDetails_Card:
		return domain.KindCard
	case *gophkeeperv1.SecretDetails_Text:
		return domain.KindText
	case *gophkeeperv1.SecretDetails_Raw:
		return domain.KindRaw
	default:
		return domain.KindUnspecified
	}
}

// convert proto SecretDetails to domain.SecretDetails
func pbDetailsToDomain(d *gophkeeperv1.SecretDetails) domain.SecretDetails {
	if d == nil {
		return domain.SecretDetails{}
	}
	out := domain.SecretDetails{
		Description: d.GetDescription(),
	}
	switch v := d.Secret.(type) {
	case *gophkeeperv1.SecretDetails_Credentials:
		out.Credentials = &domain.Credentials{
			Login:    v.Credentials.GetLogin(),
			Password: v.Credentials.GetPassword(),
		}
	case *gophkeeperv1.SecretDetails_Card:
		out.Card = &domain.Card{
			Number:           v.Card.GetNumber(),
			Holder:           v.Card.GetHolder(),
			Expires:          v.Card.GetExpires(),
			VerificationCode: v.Card.GetVerificationCode(),
		}
	case *gophkeeperv1.SecretDetails_Text:
		out.Text = &domain.Text{
			Filename: v.Text.GetFilename(),
			Content:  v.Text.GetContent(),
		}
	case *gophkeeperv1.SecretDetails_Raw:
		out.Raw = &domain.Raw{
			Filename: v.Raw.GetFilename(),
			Path:     "", // unknown at this moment
			Size:     v.Raw.GetSize(),
		}
	}
	return out
}

// convert domain.SecretDetails to proto SecretDetails
func domainDetailsToPb(d domain.SecretDetails) *gophkeeperv1.SecretDetails {
	out := &gophkeeperv1.SecretDetails{
		Description: d.Description,
	}
	switch {
	case d.Credentials != nil:
		out.Secret = &gophkeeperv1.SecretDetails_Credentials{
			Credentials: &gophkeeperv1.Credentials{
				Login:    d.Credentials.Login,
				Password: d.Credentials.Password,
			},
		}
	case d.Card != nil:
		out.Secret = &gophkeeperv1.SecretDetails_Card{
			Card: &gophkeeperv1.Card{
				Number:           d.Card.Number,
				Holder:           d.Card.Holder,
				Expires:          d.Card.Expires,
				VerificationCode: d.Card.VerificationCode,
			},
		}
	case d.Text != nil:
		out.Secret = &gophkeeperv1.SecretDetails_Text{
			Text: &gophkeeperv1.Text{
				Filename: d.Text.Filename,
				Content:  d.Text.Content,
			},
		}
	case d.Raw != nil:
		out.Secret = &gophkeeperv1.SecretDetails_Raw{
			Raw: &gophkeeperv1.Raw{
				Filename: d.Raw.Filename,
				Size:     d.Raw.Size,
			},
		}
	}
	return out
}
