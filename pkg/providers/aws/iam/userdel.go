package iam

import (
	"log"

	"github.com/aws/aws-sdk-go/service/iam"
)

func (d *IAMProvider) DelUser() {
	client := iam.New(d.Session)
	err := deleteLoginProfile(client, d.Username)
	if err != nil {
		log.Printf("[-] Delete login profile failed: %s\n", err.Error())
		return
	}
	err = detachUserPolicy(client, d.Username)
	if err != nil {
		log.Printf("[-] Remove policy from %s failed: %s\n", d.Username, err.Error())
		return
	}
	err = deleteUser(client, d.Username)
	if err != nil {
		log.Printf("[-] Delete user failed: %s\n", err.Error())
		return
	}
	log.Printf("[+] Delete user %s success!\n", d.Username)
}

func detachUserPolicy(client *iam.IAM, userName string) error {
	request := &iam.DetachUserPolicyInput{}
	request.UserName = &userName
	policyArn := "arn:aws:iam::aws:policy/AdministratorAccess"
	request.PolicyArn = &policyArn
	_, err := client.DetachUserPolicy(request)
	return err
}

func deleteLoginProfile(client *iam.IAM, userName string) error {
	request := &iam.DeleteLoginProfileInput{}
	request.UserName = &userName
	_, err := client.DeleteLoginProfile(request)
	return err
}

func deleteUser(client *iam.IAM, userName string) error {
	request := &iam.DeleteUserInput{}
	request.UserName = &userName
	_, err := client.DeleteUser(request)
	return err
}
