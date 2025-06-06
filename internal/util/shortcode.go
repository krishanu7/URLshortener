/* Estimatation:
Daily active users per second : 1000
Want to store URLs for 10 year
Total Seconds in 10 years : 365* 24 * 60 * 60 * 10 = 315360000 = 315 M
Total URLs in 10 years : 1000 * 315M = 315B URLs
>Calculation of Length of url identifers
We want to use alphanumeric characters (a-z, A-Z, 0-9) for the short code.
Total characters = 26 (lowercase) + 26 (uppercase) + 10 (digits) = 62
To calculate the length of the short code needed to accommodate 315 billion URLs, we can use the formula:
length = log_base(62, total_urls)
Where log_base(62, total_urls) is the logarithm of total_urls to the base 62.
which is approximately 6.4 so we can use 7 characters for the short code.
*/

package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateShortCode generates a random alphanumeric short code of specified length.
// func GenerateShortCode(length int) (string, error) {
// 	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
// 	shortCode := make([]byte, length)

// 	for i := range shortCode {
// 		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
// 		if err != nil {
// 			return "", fmt.Errorf("error generating random number: %w", err)
// 		}
// 		shortCode[i] = charset[num.Int64()]
// 	}

//		return string(shortCode), nil
//	}
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortCode(length int) (string, error) {
	shortCode := make([]byte, length)

	for i := range shortCode {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("error generating random number: %v", err)
		}
		shortCode[i] = charset[num.Int64()]
	}
	return string(shortCode), nil
}
