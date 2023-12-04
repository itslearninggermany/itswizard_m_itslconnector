package itswizard_m_itslconnector

import (
	"bufio"
	"bytes"
	"encoding/csv"
	itswizard_basic "github.com/itslearninggermany/itswizard_m_basic"
	imses "github.com/itslearninggermany/itswizard_m_imses"
	"io"
	"log"
	"strings"
)

type igsLengedeUsers struct {
	Id          string
	Type        string
	EMail       string
	Groups      []string
	Birthday    string
	Gender      string
	DisplayName string
	Username    string
	Password    string
	ChildId     []string
	Firstname   string
	Lastname    string
	Gruppenzug  string
	Disabled    bool
	Deleted     bool
}

type gruppeIgsLengede struct {
	Name     string
	ID       string
	ParentID string
	Level    string
}

type ItslConnectorForLengdeInput struct {
	RootGroupIdInItslearning string
	Groups                   []byte
	Itsl                     *imses.Request
	OrganisationID           uint
	InstitutionID            uint
	Users                    []byte
}

func NewItslConnectorForLengde(input ItslConnectorForLengdeInput) (*Itslconnector, error) {
	itslc := new(Itslconnector)
	x := make(map[string]bool)
	x["Student"] = true
	x["Staff"] = true
	x["Carer"] = true
	x["Guest"] = true
	itslc.profilesToSync = x
	itslc.itsl = input.Itsl
	itslc.rootGroupIdInItslearning = input.RootGroupIdInItslearning
	itslc.organisationID = input.OrganisationID
	itslc.instiutionID = input.InstitutionID
	//itslearningData
	_, err := itslc.holeAlleNutzerAusItslearning()
	if err != nil {
		return nil, err
	}
	itslc.Gruppen = PrepareGroupsIgsLengede(input.Groups, input.OrganisationID, input.InstitutionID)
	membitsl, err := itslc.hohleAlleMembershipAusItslearning()
	if err != nil {
		return nil, err
	}
	itslc.mitgliedschaftenInItslearning = membitsl
	itslc.mitgliedschaftenImFremdsytem = GetMembershipLengede(input.Users, input.Groups, input.OrganisationID, input.InstitutionID)
	itslc.alleNutzerImFremdsystem = GetUsersLengede(input.Users)
	itslc.elternKindBeziehungInItslearning = itslc.GetElternKindBeziehungInItslearning()
	itslc.elternKindBeziehungImFremdsystem = GetElternKindBeziehungInLengede(input.Users)
	return itslc, nil
}

func ReadCsvIgsUsers(userInput []byte) []igsLengedeUsers {
	var igsUsers []igsLengedeUsers
	r := csv.NewReader(bufio.NewReader(bytes.NewReader(userInput)))
	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
		}
		if record[0] == "Id" {
			continue
		}

		gr := strings.Split(record[3], "/")
		ci := strings.Split(record[9], "/")

		if record[12] == "False" {
			if record[13] == "False" {
				firstname := "NN"
				if record[10] != "" {
					firstname = record[10]
				}
				igsUsers = append(igsUsers, igsLengedeUsers{
					Id:          record[0],
					Type:        record[1],
					EMail:       record[2],
					Groups:      gr,
					Birthday:    record[4],
					Gender:      record[5],
					DisplayName: record[6],
					Username:    record[7],
					Password:    record[8],
					ChildId:     ci,
					Firstname:   firstname,
					Lastname:    record[11],
					Gruppenzug:  record[3],
				})
			}
		}
	}

	return igsUsers
}

func ReadCsvIgsGroups(groupInput []byte) []gruppeIgsLengede {
	var gruppe []gruppeIgsLengede
	r := csv.NewReader(bufio.NewReader(bytes.NewReader(groupInput)))
	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if record[0] == "name" {
			continue
		}
		if record[1] == "ID" {
			continue
		}
		gruppe = append(gruppe, gruppeIgsLengede{
			Name:     record[0],
			ID:       record[1],
			ParentID: record[2],
			Level:    record[3],
		})
	}
	return gruppe
}

func ReadCsvGroups(groupInput []byte, organisationID uint, institutionID uint) map[string]itswizard_basic.Group {
	groups := make(map[string]itswizard_basic.Group)
	gruppe := ReadCsvIgsGroups(groupInput)
	for i := 0; i < len(gruppe); i++ {
		groups[gruppe[i].Name] = itswizard_basic.Group{
			GroupSyncKey:   gruppe[i].ID,
			Name:           gruppe[i].Name,
			ParentGroupID:  gruppe[i].ParentID,
			IsCourse:       false,
			Organisation15: organisationID,
			Institution15:  institutionID,
			ToSync:         true,
		}
	}
	return groups
}

func PrepareGroupsIgsLengede(groupInput []byte, organisationID uint, institutionID uint) (gr []itswizard_basic.Group) {
	tmp := ReadCsvGroups(groupInput, organisationID, institutionID)
	for _, v := range tmp {
		gr = append(gr, v)
	}
	return
}

func GetUsersLengede(userInput []byte) map[string]itswizard_basic.Person {
	user := make(map[string]itswizard_basic.Person)
	igsUsers := ReadCsvIgsUsers(userInput)
	for i := 0; i < len(igsUsers); i++ {
		profile := "Guest"
		if igsUsers[i].Type == "S" {
			profile = "Student"
		}
		if igsUsers[i].Type == "P" {
			profile = "Staff"
		}
		if igsUsers[i].Type == "E" {
			profile = "Carer"
		}

		street1, street2 := lengedeAdressShortener(igsUsers[i].Gruppenzug)

		user[igsUsers[i].Id] = itswizard_basic.Person{
			PersonSyncKey: igsUsers[i].Id,
			FirstName:     igsUsers[i].Firstname,
			LastName:      igsUsers[i].Lastname,
			Username:      igsUsers[i].Username,
			Profile:       profile,
			Email:         igsUsers[i].EMail,
			Street1:       street1,
			Street2:       street2,
		}
	}
	return user
}

func lengedeAdressShortener(input string) (street1, street2 string) {
	tmpCounter := -1
	tmpString := ""
	var gro []string
	for _, r := range input {
		tmpCounter++
		if tmpCounter == 64 {
			tmpCounter = 0
			gro = append(gro, tmpString)
			tmpString = string(r)
		} else {
			tmpString = tmpString + string(r)
		}
	}
	gro = append(gro, tmpString)

	if len(gro) == 0 {
		return "", ""
	}
	if len(gro) == 1 {
		return gro[0], ""
	}

	return gro[0], gro[1]

}

func GetMembershipLengede(userInput []byte, groupInput []byte, organisationId uint, institutionId uint) map[string][]imses.Membership {
	membIgs := make(map[string][]imses.Membership)
	igsUsers := ReadCsvIgsUsers(userInput)
	readGroups := ReadCsvGroups(groupInput, organisationId, institutionId)

	for i := 0; i < len(igsUsers); i++ {
		for s := 0; s < len(igsUsers[i].Groups); s++ {
			tmp := membIgs[igsUsers[i].Groups[s]]
			profile := "Guest"
			if igsUsers[i].Type == "S" {
				profile = "Learner"
			}
			if igsUsers[i].Type == "P" {
				profile = "Instructor"
			}
			tmp = append(tmp, imses.Membership{
				ID:       igsUsers[i].Id + "++" + readGroups[igsUsers[i].Groups[s]].GroupSyncKey,
				GroupID:  readGroups[igsUsers[i].Groups[s]].GroupSyncKey,
				PersonID: igsUsers[i].Id,
				Profile:  profile,
			})
			membIgs[igsUsers[i].Groups[s]] = tmp
		}
	}

	membIgs2 := make(map[string][]imses.Membership)
	/*
		key wird syncID der Gruppe
	*/
	for k, v := range membIgs {
		membIgs2[readGroups[k].GroupSyncKey] = v
	}
	return membIgs2
}

func GetElternKindBeziehungInLengede(userInput []byte) map[string][]string {
	spr := make(map[string][]string)
	igsLengedeUsers := ReadCsvIgsUsers(userInput)

	for i := 0; i < len(igsLengedeUsers); i++ {
		if igsLengedeUsers[i].ChildId[0] != "" {
			tmp := []string{}
			for tu := 0; tu < len(igsLengedeUsers[i].ChildId); tu++ {
				tmp = append(tmp, igsLengedeUsers[i].ChildId[tu])
			}
			if len(tmp) > 0 {
				spr[igsLengedeUsers[i].Id] = tmp
			}
		}
	}
	return spr
}
