// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"errors"
	"fmt"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/util"
	"github.com/google/uuid"
)

const (
	MfaCountryCodeSession = "mfa_country_code"
	MfaDestSession        = "mfa_dest"
)

type SmsMfa struct {
	Config *MfaProps
}

func (mfa *SmsMfa) Initiate(ctx *context.Context, userId string) (*MfaProps, error) {
	recoveryCode := uuid.NewString()

	err := ctx.Input.CruSession.Set(MfaRecoveryCodesSession, []string{recoveryCode})
	if err != nil {
		return nil, err
	}

	mfaProps := MfaProps{
		MfaType:       mfa.Config.MfaType,
		RecoveryCodes: []string{recoveryCode},
	}
	return &mfaProps, nil
}

func (mfa *SmsMfa) SetupVerify(ctx *context.Context, passCode string) error {
	destSession := ctx.Input.CruSession.Get(MfaDestSession)
	if destSession == nil {
		return errors.New("dest session is missing")
	}
	dest := destSession.(string)

	if !util.IsEmailValid(dest) {
		countryCodeSession := ctx.Input.CruSession.Get(MfaCountryCodeSession)
		if countryCodeSession == nil {
			return errors.New("country code is missing")
		}
		countryCode := countryCodeSession.(string)

		dest, _ = util.GetE164Number(dest, countryCode)
	}

	if result := CheckVerificationCode(dest, passCode, "en"); result.Code != VerificationSuccess {
		return errors.New(result.Msg)
	}
	return nil
}

func (mfa *SmsMfa) Enable(ctx *context.Context, user *User) error {
	recoveryCodes := ctx.Input.CruSession.Get(MfaRecoveryCodesSession).([]string)
	if len(recoveryCodes) == 0 {
		return fmt.Errorf("recovery codes is missing")
	}

	columns := []string{"recovery_codes", "preferred_mfa_type"}

	user.RecoveryCodes = append(user.RecoveryCodes, recoveryCodes...)
	if user.PreferredMfaType == "" {
		user.PreferredMfaType = mfa.Config.MfaType
	}

	if mfa.Config.MfaType == SmsType {
		user.MfaPhoneEnabled = true
		columns = append(columns, "mfa_phone_enabled")

		if user.Phone == "" {
			user.Phone = ctx.Input.CruSession.Get(MfaDestSession).(string)
			user.CountryCode = ctx.Input.CruSession.Get(MfaCountryCodeSession).(string)
			columns = append(columns, "phone", "country_code")
		}
	} else if mfa.Config.MfaType == EmailType {
		user.MfaEmailEnabled = true
		columns = append(columns, "mfa_email_enabled")

		if user.Email == "" {
			user.Email = ctx.Input.CruSession.Get(MfaDestSession).(string)
			columns = append(columns, "email")
		}
	}

	_, err := UpdateUser(user.GetId(), user, columns, false)
	if err != nil {
		return err
	}

	ctx.Input.CruSession.Delete(MfaRecoveryCodesSession)
	ctx.Input.CruSession.Delete(MfaDestSession)
	ctx.Input.CruSession.Delete(MfaCountryCodeSession)

	return nil
}

func (mfa *SmsMfa) Verify(passCode string) error {
	if !util.IsEmailValid(mfa.Config.Secret) {
		mfa.Config.Secret, _ = util.GetE164Number(mfa.Config.Secret, mfa.Config.CountryCode)
	}
	if result := CheckVerificationCode(mfa.Config.Secret, passCode, "en"); result.Code != VerificationSuccess {
		return errors.New(result.Msg)
	}
	return nil
}

func NewSmsMfaUtil(config *MfaProps) *SmsMfa {
	if config == nil {
		config = &MfaProps{
			MfaType: SmsType,
		}
	}
	return &SmsMfa{
		Config: config,
	}
}

func NewEmailMfaUtil(config *MfaProps) *SmsMfa {
	if config == nil {
		config = &MfaProps{
			MfaType: EmailType,
		}
	}
	return &SmsMfa{
		Config: config,
	}
}
