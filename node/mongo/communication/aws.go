package communication

import (
	"fmt"
	"progetto-sdcc/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

/*
Crea una sessione client AWS
*/
func CreateSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials(utils.AWS_CRED_PATH, "default")})
	if err != nil {
		fmt.Println(err)
	}
	return sess
}
