package notation

// ValidationType is an enum for signature verification types such as Integrity,
// Authenticity, etc.
type ValidationType string

// ValidationAction is an enum for signature verification actions such as
// Enforced, Logged, Skipped.
type ValidationAction string

// VerificationLevel encapsulates the signature verification preset and its
// actions for each verification type
type VerificationLevel struct {
	Name        string
	Enforcement map[ValidationType]ValidationAction
}

const (
	TypeIntegrity          ValidationType = "integrity"
	TypeAuthenticity       ValidationType = "authenticity"
	TypeAuthenticTimestamp ValidationType = "authenticTimestamp"
	TypeExpiry             ValidationType = "expiry"
	TypeRevocation         ValidationType = "revocation"
)

const (
	ActionEnforce ValidationAction = "enforce"
	ActionLog     ValidationAction = "log"
	ActionSkip    ValidationAction = "skip"
)

var (
	LevelStrict = &VerificationLevel{
		Name: "strict",
		Enforcement: map[ValidationType]ValidationAction{
			TypeIntegrity:          ActionEnforce,
			TypeAuthenticity:       ActionEnforce,
			TypeAuthenticTimestamp: ActionEnforce,
			TypeExpiry:             ActionEnforce,
			TypeRevocation:         ActionEnforce,
		},
	}

	LevelPermissive = &VerificationLevel{
		Name: "permissive",
		Enforcement: map[ValidationType]ValidationAction{
			TypeIntegrity:          ActionEnforce,
			TypeAuthenticity:       ActionEnforce,
			TypeAuthenticTimestamp: ActionLog,
			TypeExpiry:             ActionLog,
			TypeRevocation:         ActionLog,
		},
	}

	LevelAudit = &VerificationLevel{
		Name: "audit",
		Enforcement: map[ValidationType]ValidationAction{
			TypeIntegrity:          ActionEnforce,
			TypeAuthenticity:       ActionLog,
			TypeAuthenticTimestamp: ActionLog,
			TypeExpiry:             ActionLog,
			TypeRevocation:         ActionLog,
		},
	}

	LevelSkip = &VerificationLevel{
		Name: "skip",
		Enforcement: map[ValidationType]ValidationAction{
			TypeIntegrity:          ActionSkip,
			TypeAuthenticity:       ActionSkip,
			TypeAuthenticTimestamp: ActionSkip,
			TypeExpiry:             ActionSkip,
			TypeRevocation:         ActionSkip,
		},
	}
)

var (
	ValidationTypes = []ValidationType{
		TypeIntegrity,
		TypeAuthenticity,
		TypeAuthenticTimestamp,
		TypeExpiry,
		TypeRevocation,
	}

	ValidationActions = []ValidationAction{
		ActionEnforce,
		ActionLog,
		ActionSkip,
	}

	VerificationLevels = []*VerificationLevel{
		LevelStrict,
		LevelPermissive,
		LevelAudit,
		LevelSkip,
	}
)

// Document represents a trustPolicy.json document
type Document struct {
	// Version of the policy document
	Version string `json:"version"`

	// TrustPolicies include each policy statement
	TrustPolicies []TrustPolicy `json:"trustPolicies"`
}

// TrustPolicy represents a policy statement in the policy document
type TrustPolicy struct {
	// Name of the policy statement
	Name string `json:"name"`

	// RegistryScopes that this policy statement affects
	RegistryScopes []string `json:"registryScopes"`

	// SignatureVerification setting for this policy statement
	SignatureVerification SignatureVerification `json:"signatureVerification"`

	// TrustStores this policy statement uses
	TrustStores []string `json:"trustStores,omitempty"`

	// TrustedIdentities this policy statement pins
	TrustedIdentities []string `json:"trustedIdentities,omitempty"`
}

// SignatureVerification represents verification configuration in a trust policy
type SignatureVerification struct {
	VerificationLevel string                              `json:"level"`
	Override          map[ValidationType]ValidationAction `json:"override,omitempty"`
}
