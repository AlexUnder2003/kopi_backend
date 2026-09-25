package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed otp.html
var files embed.FS

var otpTmpl = template.Must(template.ParseFS(files, "otp.html"))

func RenderOTP(otp string) (string, error) {
	var buf bytes.Buffer
	if err := otpTmpl.Execute(&buf, map[string]string{"OTP": otp}); err != nil {
		return "", fmt.Errorf("render otp email: %w", err)
	}
	return buf.String(), nil
}
