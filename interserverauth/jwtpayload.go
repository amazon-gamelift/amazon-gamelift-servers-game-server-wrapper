package interserverauth

type JwtPayload struct {
	Issuer         string `json:"iss"`
	ClientId       string `json:"sub"`
	Audience       string `json:"aud"`
	KeyId          string `json:"kid"`
	IssueTime      int64  `json:"iat"`
	ExpirationTime int64  `json:"exp"`
}
