package itswizard_m_itslconnector

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	itswizard_azureactivedirctory "github.com/itslearninggermany/itswizard_m_azureactivedirctory"
	itswizard_basic "github.com/itslearninggermany/itswizard_m_basic"
	imses "github.com/itslearninggermany/itswizard_m_imses"
	itswizard_msgraph "github.com/itslearninggermany/itswizard_m_msgraph"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type AadProfileSetting struct {
	Id         int               `json:"id"` // 1 ==> Groups; 2 ==> Fieldname; 3 ==> Domainname
	Groups     map[string]string `json:"groups"`
	Fieldname  string            `json:"fieldname"`
	Domainname map[string]string `json:"domainname"`
}

type AadGroupsProfile struct {
	gorm.Model
	OrganisationID uint
	InstitutionID  uint
	GroupSyncId    string
	Profile        string
}

type ItslConnectorForAzureInput struct {
	Groups                   []itswizard_basic.Group
	RootGroupIdInItslearning string
	SamlAuthGroups           []string
	AadProfileSetting        AadProfileSetting
	Itsl                     *imses.Request
	OrganisationID           uint
	InstitutionID            uint
	ProfilesToSync           []string
	TenantID                 string
	ApplicationID            string
	ClientSecret             string
}

func NewItslConnectorForAzure(input ItslConnectorForAzureInput) (*Itslconnector, error) {
	itslc := new(Itslconnector)
	itslc.itsl = input.Itsl
	itslc.rootGroupIdInItslearning = input.RootGroupIdInItslearning
	itslc.organisationID = input.OrganisationID
	itslc.instiutionID = input.InstitutionID
	itslc.Gruppen = input.Groups
	mx := make(map[string]bool)
	for i := 0; i < len(input.ProfilesToSync); i++ {
		mx[input.ProfilesToSync[i]] = true
	}
	itslc.profilesToSync = mx
	// Nutzer
	//Alle Nutzer für itslearning gesetzt
	_, err := itslc.holeAlleNutzerAusItslearning()
	if err != nil {
		return nil, err
	}
	//Alle Nutzer vom AAD
	aad, err := itswizard_azureactivedirctory.NewGraphClient(input.TenantID, input.ApplicationID, input.ClientSecret)
	if err != nil {
		return nil, err
	}

	itslc.alleNutzerImFremdsystem, err = itslc.GetAAdUsers(aad, input.SamlAuthGroups, input.AadProfileSetting, input.AadProfileSetting.Groups)
	if len(itslc.alleNutzerImFremdsystem) == 0 {
		err = errors.New("No users delivered by AAD")
	}

	//
	// Gruppen
	membitsl, err := itslc.hohleAlleMembershipAusItslearning()
	if err != nil {
		return nil, err
	}

	itslc.mitgliedschaftenInItslearning = membitsl
	membaad, err := itslc.GetAAdMemberships(input.TenantID, input.ApplicationID, input.ClientSecret, input.AadProfileSetting)
	itslc.mitgliedschaftenImFremdsytem = membaad
	return itslc, nil
}

func (p *Itslconnector) GetAAdUsers(aad *itswizard_azureactivedirctory.GraphClient, samlAuthGroups []string, aadProfileSetting AadProfileSetting, profileGroups map[string]string) (map[string]itswizard_basic.Person, error) {

	persons := make(map[string]itswizard_basic.Person)
	for s := 0; s < len(samlAuthGroups); s++ {
		fmt.Println(samlAuthGroups[s])
		/*
			if samlAuthGroups[s] == "6ef36318-3692-43aa-8980-5900482f788d" {
				fmt.Println("JO")
			}else {
				fmt.Println(samlAuthGroups[s],"6ef36318-3692-43aa-8980-5900482f788d")
			}
			user, err := aad.GetMembersOfGroup("6ef36318-3692-43aa-8980-5900482f788d")
		*/
		user, err := aad.GetMembersOfGroup(samlAuthGroups[s])
		fmt.Println(err)
		fmt.Println("HIER")
		fmt.Println(user)
		fmt.Println(len(user))
		if err != nil {
			continue
			//			return persons, err
		}
		for i := 0; i < len(user); i++ {
			fmt.Println("------")
			fmt.Println(itswizard_msgraph.UnPtrString(user[i].GivenName))
			fmt.Println("------")
			phone := ""
			if len(user[i].BusinessPhones) > 0 {
				phone = user[i].BusinessPhones[0]
			}
			tmp := itswizard_basic.Person{
				PersonSyncKey:  itswizard_msgraph.UnPtrString(user[i].ID),
				FirstName:      itswizard_msgraph.UnPtrString(user[i].GivenName),
				LastName:       itswizard_msgraph.UnPtrString(user[i].Surname),
				Email:          itswizard_msgraph.UnPtrString(user[i].Mail),
				Username:       itswizard_msgraph.UnPtrString(user[i].UserPrincipalName),
				Profile:        getAADProfile(itswizard_msgraph.UnPtrString(user[i].ID), itswizard_msgraph.UnPtrString(user[i].DisplayName), itswizard_msgraph.UnPtrString(user[i].UserPrincipalName), itswizard_msgraph.UnPtrString(user[i].JobTitle), itswizard_msgraph.UnPtrString(user[i].Department), aadProfileSetting),
				Phone:          phone,
				Mobile:         itswizard_msgraph.UnPtrString(user[i].MobilePhone),
				Street1:        itswizard_msgraph.UnPtrString(user[i].StreetAddress),
				Street2:        "",
				Postcode:       itswizard_msgraph.UnPtrString(user[i].PostalCode),
				City:           itswizard_msgraph.UnPtrString(user[i].City),
				Organisation15: p.organisationID,
				Institution15:  p.instiutionID,
			}
			persons[tmp.PersonSyncKey] = tmp
		}
	}
	fmt.Println("FERTIG")
	return persons, nil
}

func getAADProfile(personAzureID string, displayname string, userPrincipalName string, userJobTitle string, userDepartment string, aadProfileSetting AadProfileSetting) string {
	/*
		Id         int               `json:"id"` // 1 ==> Groups; 2 ==> Fieldname; 3 ==> Domainname
		Groups     map[string]string `json:"groups"`  map[azureID]string
		Fieldname  string            `json:"fieldname"`
		Domainname map[string]string `json:"domainname"` map[profile]domainname
	*/
	switch aadProfileSetting.Id {
	case 1:
		/*
			Schauen ob person in einer der Gruppe ist
		*/
		profile := aadProfileSetting.Groups[personAzureID]
		if profile == "" {
			log.Println("Problem with the profile group user", displayname, " not in any of the groups")
			return "Guest"
		} else {
			return profile
		}
	case 2:
		if aadProfileSetting.Fieldname == "Jobtitle" {
			if userJobTitle == "" {
				return "Guest"
			} else {
				if aadProfileSetting.Groups[userJobTitle] == "" {
					return "Guest"
				} else {
					return aadProfileSetting.Groups[userJobTitle]
				}
			}
		} else if aadProfileSetting.Fieldname == "department" {
			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++")
			fmt.Println("Department")
			fmt.Println(userDepartment)
			fmt.Println(aadProfileSetting.Groups)
			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++")
			if userDepartment == "" {
				return "Guest"
			} else {
				// AUF KSB asugerichtet
				if userDepartment == "Lehrperson" {
					return "Staff"
				}
				if userDepartment == "Schueler" {
					return "Student"
				} else {
					return "Guest"
				}
			}
		} else {
			log.Println("Fieldname not in script for", displayname)
			return "Guest"
		}
	case 3:
		for k, v := range aadProfileSetting.Domainname {
			if strings.HasSuffix(userPrincipalName, k) {
				return v
			}
		}
		log.Println("Domainname not found for", userPrincipalName)
		return "Guest"
	default:
		log.Println("Problem with getAADProfile. It is not case 1,2 or 3")
		return "Guest"
	}
}

func (p *Itslconnector) GetAAdMemberships(tenantID, applicationID, secretID string, aadProfileSetting AadProfileSetting) (map[string][]imses.Membership, error) {
	groupMembers := make(map[string][]imses.Membership)

	for g := 0; g < len(p.Gruppen); g++ {
		userIds, err := getMemberIdsHelpFunc(tenantID, applicationID, secretID, p.Gruppen[g].AzureGroupID)
		fmt.Println("Gruppe:", p.Gruppen[g], "Mitglieder:", len(userIds))
		if err != nil {
			return nil, err
		}
		var tmp []imses.Membership
		for i := 0; i < len(userIds); i++ {

			profile := "Guest"
			if p.alleNutzerImFremdsystem[userIds[i]].Profile == "Staff" {
				profile = "Instructor"
			}
			if p.alleNutzerImFremdsystem[userIds[i]].Profile == "Student" {
				profile = "Learner"
			}
			tmp = append(tmp, imses.Membership{
				ID:       p.Gruppen[g].GroupSyncKey + p.alleNutzerImFremdsystem[userIds[i]].PersonSyncKey,
				GroupID:  p.Gruppen[g].GroupSyncKey,
				PersonID: p.alleNutzerImFremdsystem[userIds[i]].PersonSyncKey,
				Profile:  profile,
			})

		}
		groupMembers[p.Gruppen[g].GroupSyncKey] = tmp
	}
	return groupMembers, nil
}

func getMemberIdsHelpFunc(TenantID, ClientID, ClientSecret, ID string) ([]string, error) {
	var allIds []string

	resource := fmt.Sprintf("/%v/oauth2/token", TenantID)
	data := url.Values{}
	data.Add("grant_type", "client_credentials")
	data.Add("client_id", ClientID)
	data.Add("client_secret", ClientSecret)
	data.Add("resource", "https://graph.microsoft.com")

	u, err := url.ParseRequestURI("https://login.microsoftonline.com")
	if err != nil {
		fmt.Println("Unable to parse URI: %v", err)
	}

	u.Path = resource
	req, err := http.NewRequest("POST", u.String(), bytes.NewBufferString(data.Encode()))

	if err != nil {
		fmt.Println("HTTP Request Error: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("HTTP response error: %v of http.Request: %v", err, req.URL)
	}

	defer resp.Body.Close() // close body when func returns

	body, err := ioutil.ReadAll(resp.Body) // read body first to append it to the error (if any)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Hint: this will mostly be the case if the tenant ID can not be found, the Application ID can not be found or the clientSecret is incorrect.
		// The cause will be described in the body, hence we have to return the body too for proper error-analysis
		fmt.Println("StatusCode is not OK: %v. Body: %v ", resp.StatusCode, string(body))
	}

	if err != nil {
		fmt.Println("HTTP response read error: %v of http.Request: %v", err, req.URL)
	}
	var token Token
	json.Unmarshal(body, &token)

	resource2 := "/groups/" + ID + "/members"

	reqURL, err := url.ParseRequestURI("https://graph.microsoft.com")
	if err != nil {
		fmt.Println("Unable to parse URI %v: %v", "https://graph.microsoft.com", err)
	}
	reqURL.Path = "/" + "v1.0" + resource2

	req2, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		fmt.Println("HTTP request error: %v", err)
	}

	req2.Header.Add("Content-Type", "application/json")
	tok := fmt.Sprintf("%v %v", token.TokenType, token.AccessToken)
	req2.Header.Add("Authorization", tok)

	getParams := url.Values{}
	getParams.Add("$top", strconv.Itoa(999))
	req2.URL.RawQuery = getParams.Encode() // set query parameters

	httpClient2 := &http.Client{
		Timeout: time.Second * 10,
	}
	resp2, err := httpClient2.Do(req2)
	if err != nil {
		fmt.Println("HTTP response error: %v of http.Request: %v", err, req.URL)
	}

	defer resp2.Body.Close() // close body when func returns

	body2, err := ioutil.ReadAll(resp2.Body) // read body first to append it to the error (if any)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Hint: this will mostly be the case if the tenant ID can not be found, the Application ID can not be found or the clientSecret is incorrect.
		// The cause will be described in the body, hence we have to return the body too for proper error-analysis
		fmt.Println("StatusCode is not OK: %v. Body: %v ", resp.StatusCode, string(body))
	}

	if err != nil {
		fmt.Println("HTTP response read error: %v of http.Request: %v", err, req2.URL)
	}

	var groupMembersOutput GroupMembersOutput
	json.Unmarshal(body2, &groupMembersOutput)
	for hu := 0; hu < len(groupMembersOutput.Value); hu++ {
		allIds = append(allIds, groupMembersOutput.Value[hu].ID)
	}

	if strings.Compare(groupMembersOutput.OdataNextLink, "") > 0 {
		for {
			lastLink := groupMembersOutput.OdataNextLink
			req, err := http.NewRequest("GET", groupMembersOutput.OdataNextLink, nil)
			if err != nil {
				fmt.Println(err)
			}

			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", tok)

			httpClient := &http.Client{
				Timeout: time.Second * 10,
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				fmt.Println(err)
			}

			defer resp.Body.Close() // close body when func returns

			body, err := ioutil.ReadAll(resp.Body)

			json.Unmarshal(body, &groupMembersOutput)
			for hu := 0; hu < len(groupMembersOutput.Value); hu++ {
				allIds = append(allIds, groupMembersOutput.Value[hu].ID)
			}
			if groupMembersOutput.OdataNextLink == lastLink {
				break
			}
		}
	}
	return allIds, nil
}

type Token struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
}

type GroupMembersOutput struct {
	OdataContext  string `json:"@odata.context"`
	OdataNextLink string `json:"@odata.nextLink"`
	Value         []struct {
		OdataType         string      `json:"@odata.type"`
		ID                string      `json:"id"`
		BusinessPhones    []string    `json:"businessPhones"`
		DisplayName       string      `json:"displayName"`
		GivenName         string      `json:"givenName"`
		JobTitle          string      `json:"jobTitle"`
		Mail              string      `json:"mail"`
		MobilePhone       interface{} `json:"mobilePhone"`
		OfficeLocation    interface{} `json:"officeLocation"`
		PreferredLanguage interface{} `json:"preferredLanguage"`
		Surname           string      `json:"surname"`
		UserPrincipalName string      `json:"userPrincipalName"`
	} `json:"value"`
}
