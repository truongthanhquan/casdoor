// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type SignupItem struct {
	Name     string `json:"name"`
	Visible  bool   `json:"visible"`
	Required bool   `json:"required"`
	Prompted bool   `json:"prompted"`
	Rule     string `json:"rule"`
}

type Application struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName         string          `xorm:"varchar(100)" json:"displayName"`
	Logo                string          `xorm:"varchar(100)" json:"logo"`
	HomepageUrl         string          `xorm:"varchar(100)" json:"homepageUrl"`
	Description         string          `xorm:"varchar(100)" json:"description"`
	Organization        string          `xorm:"varchar(100)" json:"organization"`
	Cert                string          `xorm:"varchar(100)" json:"cert"`
	EnablePassword      bool            `json:"enablePassword"`
	EnableSignUp        bool            `json:"enableSignUp"`
	EnableSigninSession bool            `json:"enableSigninSession"`
	EnableAutoSignin    bool            `json:"enableAutoSignin"`
	EnableCodeSignin    bool            `json:"enableCodeSignin"`
	EnableSamlCompress  bool            `json:"enableSamlCompress"`
	EnableWebAuthn      bool            `json:"enableWebAuthn"`
	EnableLinkWithEmail bool            `json:"enableLinkWithEmail"`
	SamlReplyUrl        string          `xorm:"varchar(100)" json:"samlReplyUrl"`
	Providers           []*ProviderItem `xorm:"mediumtext" json:"providers"`
	SignupItems         []*SignupItem   `xorm:"varchar(1000)" json:"signupItems"`
	GrantTypes          []string        `xorm:"varchar(1000)" json:"grantTypes"`
	OrganizationObj     *Organization   `xorm:"-" json:"organizationObj"`

	ClientId             string     `xorm:"varchar(100)" json:"clientId"`
	ClientSecret         string     `xorm:"varchar(100)" json:"clientSecret"`
	RedirectUris         []string   `xorm:"varchar(1000)" json:"redirectUris"`
	TokenFormat          string     `xorm:"varchar(100)" json:"tokenFormat"`
	ExpireInHours        int        `json:"expireInHours"`
	RefreshExpireInHours int        `json:"refreshExpireInHours"`
	SignupUrl            string     `xorm:"varchar(200)" json:"signupUrl"`
	SigninUrl            string     `xorm:"varchar(200)" json:"signinUrl"`
	ForgetUrl            string     `xorm:"varchar(200)" json:"forgetUrl"`
	AffiliationUrl       string     `xorm:"varchar(100)" json:"affiliationUrl"`
	TermsOfUse           string     `xorm:"varchar(100)" json:"termsOfUse"`
	SignupHtml           string     `xorm:"mediumtext" json:"signupHtml"`
	SigninHtml           string     `xorm:"mediumtext" json:"signinHtml"`
	ThemeData            *ThemeData `xorm:"json" json:"themeData"`
	FormCss              string     `xorm:"text" json:"formCss"`
	FormCssMobile        string     `xorm:"text" json:"formCssMobile"`
	FormOffset           int        `json:"formOffset"`
	FormSideHtml         string     `xorm:"mediumtext" json:"formSideHtml"`
	FormBackgroundUrl    string     `xorm:"varchar(200)" json:"formBackgroundUrl"`
}

func GetApplicationCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Application{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetOrganizationApplicationCount(owner, Organization, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Application{Organization: Organization})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetApplications(owner string) []*Application {
	applications := []*Application{}
	err := adapter.Engine.Desc("created_time").Find(&applications, &Application{Owner: owner})
	if err != nil {
		panic(err)
	}

	return applications
}

func GetOrganizationApplications(owner string, organization string) []*Application {
	applications := []*Application{}
	err := adapter.Engine.Desc("created_time").Find(&applications, &Application{Organization: organization})
	if err != nil {
		panic(err)
	}

	return applications
}

func GetPaginationApplications(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Application {
	var applications []*Application
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&applications)
	if err != nil {
		panic(err)
	}

	return applications
}

func GetPaginationOrganizationApplications(owner, organization string, offset, limit int, field, value, sortField, sortOrder string) []*Application {
	applications := []*Application{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&applications, &Application{Organization: organization})
	if err != nil {
		panic(err)
	}

	return applications
}

func getProviderMap(owner string) map[string]*Provider {
	providers := GetProviders(owner)
	m := map[string]*Provider{}
	for _, provider := range providers {
		// Get QRCode only once
		if provider.Type == "WeChat" && provider.DisableSsl == true && provider.Content == "" {
			provider.Content, _ = idp.GetWechatOfficialAccountQRCode(provider.ClientId2, provider.ClientSecret2)
			UpdateProvider(provider.Owner+"/"+provider.Name, provider)
		}

		m[provider.Name] = GetMaskedProvider(provider)
	}
	return m
}

func extendApplicationWithProviders(application *Application) {
	m := getProviderMap(application.Organization)
	for _, providerItem := range application.Providers {
		if provider, ok := m[providerItem.Name]; ok {
			providerItem.Provider = provider
		}
	}
}

func extendApplicationWithOrg(application *Application) {
	organization := getOrganization(application.Owner, application.Organization)
	application.OrganizationObj = organization
}

func getApplication(owner string, name string) *Application {
	if owner == "" || name == "" {
		return nil
	}

	application := Application{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		extendApplicationWithProviders(&application)
		extendApplicationWithOrg(&application)
		return &application
	} else {
		return nil
	}
}

func GetApplicationByOrganizationName(organization string) *Application {
	application := Application{}
	existed, err := adapter.Engine.Where("organization=?", organization).Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		extendApplicationWithProviders(&application)
		extendApplicationWithOrg(&application)
		return &application
	} else {
		return nil
	}
}

func GetApplicationByUser(user *User) *Application {
	if user.SignupApplication != "" {
		return getApplication("admin", user.SignupApplication)
	} else {
		return GetApplicationByOrganizationName(user.Owner)
	}
}

func GetApplicationByUserId(userId string) (*Application, *User) {
	var application *Application

	owner, name := util.GetOwnerAndNameFromId(userId)
	if owner == "app" {
		application = getApplication("admin", name)
		return application, nil
	}

	user := GetUser(userId)
	application = GetApplicationByUser(user)

	return application, user
}

func GetApplicationByClientId(clientId string) *Application {
	application := Application{}
	existed, err := adapter.Engine.Where("client_id=?", clientId).Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		extendApplicationWithProviders(&application)
		extendApplicationWithOrg(&application)
		return &application
	} else {
		return nil
	}
}

func GetApplication(id string) *Application {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getApplication(owner, name)
}

func GetMaskedApplication(application *Application, userId string) *Application {
	if isUserIdGlobalAdmin(userId) {
		return application
	}

	if application == nil {
		return nil
	}

	if application.ClientSecret != "" {
		application.ClientSecret = "***"
	}

	if application.OrganizationObj != nil {
		if application.OrganizationObj.MasterPassword != "" {
			application.OrganizationObj.MasterPassword = "***"
		}
		if application.OrganizationObj.PasswordType != "" {
			application.OrganizationObj.PasswordType = "***"
		}
		if application.OrganizationObj.PasswordSalt != "" {
			application.OrganizationObj.PasswordSalt = "***"
		}
	}
	return application
}

func GetMaskedApplications(applications []*Application, userId string) []*Application {
	if isUserIdGlobalAdmin(userId) {
		return applications
	}

	for _, application := range applications {
		application = GetMaskedApplication(application, userId)
	}
	return applications
}

func UpdateApplication(id string, application *Application) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	oldApplication := getApplication(owner, name)
	if oldApplication == nil {
		return false
	}

	if name == "app-built-in" {
		application.Name = name
	}

	if name != application.Name {
		err := applicationChangeTrigger(name, application.Name)
		if err != nil {
			return false
		}
	}

	if oldApplication.ClientId != application.ClientId && GetApplicationByClientId(application.ClientId) != nil {
		return false
	}

	for _, providerItem := range application.Providers {
		providerItem.Provider = nil
	}

	session := adapter.Engine.ID(core.PK{owner, name}).AllCols()
	if application.ClientSecret == "***" {
		session.Omit("client_secret")
	}
	affected, err := session.Update(application)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddApplication(application *Application) bool {
	if application.ClientId == "" {
		application.ClientId = util.GenerateClientId()
	}
	if application.ClientSecret == "" {
		application.ClientSecret = util.GenerateClientSecret()
	}
	if GetApplicationByClientId(application.ClientId) != nil {
		return false
	}
	for _, providerItem := range application.Providers {
		providerItem.Provider = nil
	}

	affected, err := adapter.Engine.Insert(application)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteApplication(application *Application) bool {
	if application.Name == "app-built-in" {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{application.Owner, application.Name}).Delete(&Application{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (application *Application) GetId() string {
	return fmt.Sprintf("%s/%s", application.Owner, application.Name)
}

func (application *Application) IsRedirectUriValid(redirectUri string) bool {
	isValid := false
	for _, targetUri := range application.RedirectUris {
		targetUriRegex := regexp.MustCompile(targetUri)
		if targetUriRegex.MatchString(redirectUri) || strings.Contains(redirectUri, targetUri) {
			isValid = true
			break
		}
	}
	return isValid
}

func IsOriginAllowed(origin string) bool {
	applications := GetApplications("")
	for _, application := range applications {
		if application.IsRedirectUriValid(origin) {
			return true
		}
	}
	return false
}

func getApplicationMap(organization string) map[string]*Application {
	applications := GetOrganizationApplications("admin", organization)

	applicationMap := make(map[string]*Application)
	for _, application := range applications {
		applicationMap[application.Name] = application
	}

	return applicationMap
}

func ExtendManagedAccountsWithUser(user *User) *User {
	if user.ManagedAccounts == nil || len(user.ManagedAccounts) == 0 {
		return user
	}

	applicationMap := getApplicationMap(user.Owner)

	var managedAccounts []ManagedAccount
	for _, managedAccount := range user.ManagedAccounts {
		application := applicationMap[managedAccount.Application]
		if application != nil {
			managedAccount.SigninUrl = application.SigninUrl
			managedAccounts = append(managedAccounts, managedAccount)
		}
	}
	user.ManagedAccounts = managedAccounts

	return user
}

func applicationChangeTrigger(oldName string, newName string) error {
	session := adapter.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	organization := new(Organization)
	organization.DefaultApplication = newName
	_, err = session.Where("default_application=?", oldName).Update(organization)
	if err != nil {
		return err
	}

	user := new(User)
	user.SignupApplication = newName
	_, err = session.Where("signup_application=?", oldName).Update(user)
	if err != nil {
		return err
	}

	resource := new(Resource)
	resource.Application = newName
	_, err = session.Where("application=?", oldName).Update(resource)
	if err != nil {
		return err
	}

	var permissions []*Permission
	err = adapter.Engine.Find(&permissions)
	if err != nil {
		return err
	}
	for i := 0; i < len(permissions); i++ {
		permissionResoureces := permissions[i].Resources
		for j := 0; j < len(permissionResoureces); j++ {
			if permissionResoureces[j] == oldName {
				permissionResoureces[j] = newName
			}
		}
		permissions[i].Resources = permissionResoureces
		_, err = session.Where("name=?", permissions[i].Name).Update(permissions[i])
		if err != nil {
			return err
		}
	}

	return session.Commit()
}
