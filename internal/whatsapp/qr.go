package whatsapp

import (
	"fmt"
	"time"
)

func GenerateQRCode(accountID string) string {
	return fmt.Sprintf("qr:%s:%d", NormalizeAccountID(accountID), time.Now().Unix())
}
