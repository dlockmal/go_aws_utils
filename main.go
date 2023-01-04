package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

var client *iam.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client = iam.NewFromConfig(cfg)
}

func main() {
	parms := &iam.ListUsersInput{
		MaxItems: aws.Int32(100),
	}

	result, err := client.ListUsers(context.TODO(), parms)

	if err != nil {
		fmt.Println("Error calling iam: ", err)
		return
	}

	count := len(result.Users)
	fmt.Println("Users: ", count)

	for _, user := range result.Users {
		fmt.Printf("\t%v\n", *user.UserName)

		params := &iam.ListAccessKeysInput{
			UserName: aws.String(*user.UserName),
		}
		info, err := client.ListAccessKeys(context.TODO(), params)
		if err != nil {
			fmt.Println("Error calling GetUserName: ", err)
			return
		}
		for _, k := range info.AccessKeyMetadata {
			fmt.Printf("\t%v\t%v\n", *k.AccessKeyId, *k.CreateDate)
		}
	}

}
