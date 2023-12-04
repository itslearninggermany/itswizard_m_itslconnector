package itswizard_m_itslconnector

import (
	"encoding/json"
	"fmt"
	itswizard_basic "github.com/itslearninggermany/itswizard_m_basic"
	imses "github.com/itslearninggermany/itswizard_m_imses"
	"github.com/jinzhu/gorm"
	"log"
)

type SyncCache struct {
	UserToImport                     map[string]itswizard_basic.Person
	UserToUpdate                     map[string]UserUpdate
	UserToDelete                     map[string]itswizard_basic.Person
	MembershipsToImport              map[string][]imses.Membership
	MembershipsToDelete              map[string][]imses.Membership
	itsl                             *imses.Request
	Groups                           []itswizard_basic.Group
	RootGroup                        string
	AlleNutzerImFremdsystem          map[string]itswizard_basic.Person
	AlleNutzerInItslearning          map[string]itswizard_basic.Person
	ElternKindBeziehungImFremdsystem map[string][]string
}

func ItslConnectorFromJson(b []byte, itsl *imses.Request) (itsSync SyncCache, err error) {
	err = json.Unmarshal(b, &itsSync)
	if err != nil {
		return
	}
	itsSync.itsl = itsl
	return
}

func (p *Itslconnector) GetSyncCache() SyncCache {
	x := SyncCache{
		UserToImport:                     p.GetUserToImport(),
		UserToUpdate:                     p.GetUsersToUpdate(),
		UserToDelete:                     p.GetUserToDelete(),
		MembershipsToImport:              p.GetMembershipToImport(),
		MembershipsToDelete:              p.GetMembershipToDelete(),
		itsl:                             p.itsl,
		Groups:                           p.Gruppen,
		RootGroup:                        p.rootGroupIdInItslearning,
		AlleNutzerImFremdsystem:          p.alleNutzerImFremdsystem,
		AlleNutzerInItslearning:          p.GetItslearningUsers(),
		ElternKindBeziehungImFremdsystem: p.elternKindBeziehungImFremdsystem,
	}
	return x
}

type ChangeLog struct {
	gorm.Model
	UserOrGroup      string
	NewPerson        bool
	DeltePerson      bool
	UdpatePerson     bool
	GroupImport      bool
	GroupDelete      bool
	MembershipImport bool
	MembershipDelete bool
	PSR              bool
	Error            string `gorm:"type:longtext"`
}

func (p *SyncCache) Run(organisationID, institutionID uint, psr bool) []ChangeLog {
	var changeLogs []ChangeLog
	var gruppenZuImportieren []itswizard_basic.Group
	var gruppenZuLoeschen []itswizard_basic.Group
	for i := 0; i < len(p.Groups); i++ {
		if p.Groups[i].ToSync {
			gruppenZuImportieren = append(gruppenZuImportieren, p.Groups[i])
		} else {
			gruppenZuLoeschen = append(gruppenZuLoeschen, p.Groups[i])
		}
	}

	// Gruppen
	for i := 0; i < len(gruppenZuImportieren); i++ {
		kursgruppeErstellt := false
		if gruppenZuImportieren[i].IsCourse {
			if !kursgruppeErstellt {
				resp, err := p.itsl.CreateGroup(itswizard_basic.DbGroup15{
					ID:            p.RootGroup + "kursGruppe",
					SyncID:        p.RootGroup + "kursGruppe",
					Name:          "Kurse",
					ParentGroupID: p.RootGroup,
					Level:         0,
					IsCourse:      false,
				}, false)
				if err != nil {
					changeLogs = append(changeLogs, ChangeLog{
						UserOrGroup: p.RootGroup + "kursGruppe",
						GroupImport: true,
						Error:       resp,
					})
				} else {
					changeLogs = append(changeLogs, ChangeLog{
						UserOrGroup: p.RootGroup + "kursGruppe",
						GroupImport: true,
					})
					kursgruppeErstellt = true
				}
			}
			_, err, _ := p.itsl.ReadMembershipsForGroup(gruppenZuImportieren[i].GroupSyncKey)
			if err != nil {
				resp, err := p.itsl.CreateCourse(itswizard_basic.DbGroup15{
					SyncID:        gruppenZuImportieren[i].GroupSyncKey,
					Name:          gruppenZuImportieren[i].Name,
					ParentGroupID: p.RootGroup + "kursGruppe",
					Level:         0,
				})
				if err != nil {
					changeLogs = append(changeLogs, ChangeLog{
						UserOrGroup: gruppenZuImportieren[i].GroupSyncKey,
						GroupImport: true,
						Error:       resp,
					})
				} else {
					changeLogs = append(changeLogs, ChangeLog{
						UserOrGroup: gruppenZuImportieren[i].GroupSyncKey,
						GroupImport: true,
					})
				}
			}
		} else {
			_, err, _ := p.itsl.ReadMembershipsForGroup(gruppenZuImportieren[i].GroupSyncKey)
			if err != nil {
				resp, err := p.itsl.CreateGroup(itswizard_basic.DbGroup15{
					ID:                  gruppenZuImportieren[i].GroupSyncKey,
					SyncID:              gruppenZuImportieren[i].GroupSyncKey,
					AzureGroupID:        gruppenZuImportieren[i].GroupSyncKey,
					UniventionGroupName: "",
					Name:                gruppenZuImportieren[i].Name,
					ParentGroupID:       gruppenZuImportieren[i].ParentGroupID,
					Level:               0,
					IsCourse:            false,
					DbInstitution15ID:   institutionID,
					DbOrganisation15ID:  organisationID,
				}, false)
				if err != nil {
					changeLogs = append(changeLogs, ChangeLog{
						UserOrGroup: gruppenZuImportieren[i].GroupSyncKey,
						GroupImport: true,
						Error:       resp,
					})
				} else {
					changeLogs = append(changeLogs, ChangeLog{
						UserOrGroup: gruppenZuImportieren[i].GroupSyncKey,
						GroupImport: true,
					})
				}
			}
		}

	}
	for i := 0; i < len(gruppenZuLoeschen); i++ {
		resp, err := p.itsl.DeleteGroup(gruppenZuLoeschen[i].GroupSyncKey)
		if err != nil {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: gruppenZuLoeschen[i].GroupSyncKey,
				GroupDelete: true,
				Error:       resp,
			})
		} else {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: gruppenZuLoeschen[i].GroupSyncKey,
				GroupDelete: true,
			})
		}

	}

	// Personen import und update
	for _, v := range p.UserToImport {
		log.Println("Sync", v.Username)
		resp, err := p.itsl.CreatePerson(itswizard_basic.DbPerson15{
			SyncPersonKey: v.PersonSyncKey,
			FirstName:     v.FirstName,
			LastName:      v.LastName,
			Username:      v.Username,
			Profile:       v.Profile,
			Email:         v.Email,
			Phone:         v.Phone,
			Mobile:        v.Mobile,
			Street1:       v.Street1,
			Street2:       v.Street2,
			Postcode:      v.Postcode,
			City:          v.City,
		})
		if err != nil {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: v.PersonSyncKey,
				NewPerson:   true,
				Error:       resp,
			})
		} else {
			profile := "Guest"
			if v.Profile == "Student" {
				profile = "Learner"
			}
			if v.Profile == "Staff" {
				profile = "Instructor"
			}
			resp, err := p.itsl.CreateMembership(p.RootGroup, v.PersonSyncKey, profile)
			if err != nil {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      p.RootGroup + "::" + v.PersonSyncKey,
					MembershipImport: true,
					Error:            resp,
				})
			} else {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup: v.PersonSyncKey,
					NewPerson:   true,
				})
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      p.RootGroup + "::" + v.PersonSyncKey,
					MembershipImport: true,
					NewPerson:        true,
				})
			}
		}
	}
	for _, v := range p.UserToUpdate {
		log.Println("Sync", v.UsersNew.Username)
		resp, err := p.itsl.CreatePerson(itswizard_basic.DbPerson15{
			SyncPersonKey: v.UsersNew.PersonSyncKey,
			FirstName:     v.UsersNew.FirstName,
			LastName:      v.UsersNew.LastName,
			Username:      v.UsersNew.Username,
			Profile:       v.UsersNew.Profile,
			Email:         v.UsersNew.Email,
			Phone:         v.UsersNew.Phone,
			Mobile:        v.UsersNew.Mobile,
			Street1:       v.UsersNew.Street1,
			Street2:       v.UsersNew.Street2,
			Postcode:      v.UsersNew.Postcode,
			City:          v.UsersNew.City,
		})
		if err != nil {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: v.UsersNew.PersonSyncKey,
				NewPerson:   true,
				Error:       resp,
			})
		} else {
			profile := "Guest"
			if v.UsersNew.Profile == "Student" {
				profile = "Learner"
			}
			if v.UsersNew.Profile == "Staff" {
				profile = "Instructor"
			}
			resp, err := p.itsl.CreateMembership(p.RootGroup, v.UsersNew.PersonSyncKey, profile)
			if err != nil {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      p.RootGroup + "::" + v.UsersNew.PersonSyncKey,
					MembershipImport: true,
					Error:            resp,
				})
			} else {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup: v.UsersNew.PersonSyncKey,
					NewPerson:   true,
				})
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      p.RootGroup + "::" + v.UsersNew.PersonSyncKey,
					MembershipImport: true,
					NewPerson:        true,
				})
			}
		}
	}
	for k, _ := range p.UserToDelete {
		log.Println("Sync", k)
		resp, err := p.itsl.DeletePerson(k)
		if err != nil {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: k,
				DeltePerson: true,
				Error:       resp,
			})
		} else {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: k,
				DeltePerson: true,
			})
		}
	}

	//Membership
	for _, v := range p.MembershipsToImport {
		log.Println("Membership")
		for i := 0; i < len(v); i++ {
			resp, err := p.itsl.CreateMembership(v[i].GroupID, v[i].PersonID, v[i].Profile)
			if err != nil {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      v[i].GroupID + "::" + v[i].PersonID,
					MembershipImport: true,
					Error:            resp,
				})
			} else {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      v[i].GroupID + "::" + v[i].PersonID,
					MembershipImport: true,
				})
			}
		}
	}
	for _, v := range p.MembershipsToDelete {
		log.Println("Membership")
		for i := 0; i < len(v); i++ {
			resp, err := p.itsl.DeleteMembership(v[i].ID)
			if err != nil {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      v[i].GroupID + "::" + v[i].PersonID,
					MembershipDelete: true,
					Error:            resp,
				})
			} else {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      v[i].GroupID + "::" + v[i].PersonID,
					MembershipDelete: true,
				})
			}
		}
	}

	return changeLogs
}

func (p *SyncCache) RunWithAllGroupsWithParentID(organisationID, institutionID uint, psr bool) (out []ChangeLog) {
	log.Println("Groups")
	for _, v := range p.Groups {
		resp, err := p.itsl.CreateGroup(itswizard_basic.DbGroup15{
			SyncID:        v.GroupSyncKey,
			Name:          v.Name,
			ParentGroupID: v.ParentGroupID,
			IsCourse:      false,
		}, false)
		e := ""
		if err != nil {
			e = resp
		}
		out = append(out, ChangeLog{
			UserOrGroup:      v.GroupSyncKey + " " + v.Name,
			GroupImport:      false,
			GroupDelete:      true,
			MembershipImport: false,
			MembershipDelete: false,
			PSR:              false,
			Error:            e,
		})
	}

	log.Println("User")
	l := p.OnlyUserUpdateImportAndDelete(organisationID, institutionID)
	for _, v := range l {
		out = append(out, v)
	}

	if psr {
		log.Println("PSR")
		l := p.RunPSR()
		for _, v := range l {
			out = append(out, v)
		}
	}

	log.Println("Membership Delete")
	for _, tmp := range p.MembershipsToDelete {
		for _, v := range tmp {
			resp, err := p.itsl.DeleteMembership(v.ID)
			e := ""
			if err != nil {
				e = resp
			}
			out = append(out, ChangeLog{
				UserOrGroup:      v.PersonID + "++" + v.GroupID,
				GroupImport:      false,
				GroupDelete:      false,
				MembershipImport: false,
				MembershipDelete: true,
				PSR:              false,
				Error:            e,
			})
		}
	}
	log.Println("Membership Import")

	for _, tmp := range p.MembershipsToImport {
		for _, v := range tmp {
			resp, err := p.itsl.CreateMembership(v.GroupID, v.PersonID, v.Profile)
			e := ""
			if err != nil {
				e = resp
			}
			out = append(out, ChangeLog{
				UserOrGroup:      v.PersonID + " " + v.GroupID,
				GroupImport:      false,
				GroupDelete:      false,
				MembershipImport: false,
				MembershipDelete: true,
				PSR:              false,
				Error:            e,
			})
		}
	}

	return
}

func (p *SyncCache) OnlyUserUpdateImportAndDelete(organisationID, institutionID uint) []ChangeLog {
	var changeLogs []ChangeLog
	// Personen import und update fi
	for _, v := range p.UserToImport {
		resp, err := p.itsl.CreatePerson(itswizard_basic.DbPerson15{
			SyncPersonKey: v.PersonSyncKey,
			FirstName:     v.FirstName,
			LastName:      v.LastName,
			Username:      v.Username,
			Profile:       v.Profile,
			Email:         v.Email,
			Phone:         v.Phone,
			Mobile:        v.Mobile,
			Street1:       v.Street1,
			Street2:       v.Street2,
			Postcode:      v.Postcode,
			City:          v.City,
		})
		if err != nil {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: v.PersonSyncKey,
				NewPerson:   true,
				Error:       resp,
			})
		} else {
			profile := "Guest"
			if v.Profile == "Student" {
				profile = "Learner"
			}
			if v.Profile == "Staff" {
				profile = "Instructor"
			}
			resp, err := p.itsl.CreateMembership(p.RootGroup, v.PersonSyncKey, profile)
			if err != nil {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      p.RootGroup + "::" + v.PersonSyncKey,
					MembershipImport: true,
					Error:            resp,
				})
			} else {
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup: v.PersonSyncKey,
					NewPerson:   true,
				})
				changeLogs = append(changeLogs, ChangeLog{
					UserOrGroup:      p.RootGroup + "::" + v.PersonSyncKey,
					MembershipImport: true,
					NewPerson:        true,
				})
			}
		}
	}
	for _, v := range p.UserToUpdate {
		resp, err := p.itsl.CreatePerson(itswizard_basic.DbPerson15{
			SyncPersonKey: v.UsersNew.PersonSyncKey,
			FirstName:     v.UsersNew.FirstName,
			LastName:      v.UsersNew.LastName,
			Username:      v.UsersNew.Username,
			Profile:       v.UsersNew.Profile,
			Email:         v.UsersNew.Email,
			Phone:         v.UsersNew.Phone,
			Mobile:        v.UsersNew.Mobile,
			Street1:       v.UsersNew.Street1,
			Street2:       v.UsersNew.Street2,
			Postcode:      v.UsersNew.Postcode,
			City:          v.UsersNew.City,
		})
		if err != nil {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: v.UsersNew.PersonSyncKey,
				NewPerson:   true,
				Error:       resp,
			})
		} else {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: v.UsersNew.PersonSyncKey,
				NewPerson:   true,
			})
		}
	}
	for k, _ := range p.UserToDelete {
		resp, err := p.itsl.DeletePerson(k)
		if err != nil {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: k,
				DeltePerson: true,
				Error:       resp,
			})
		} else {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: k,
				DeltePerson: true,
			})
		}
	}
	return changeLogs
}

func (p *SyncCache) RunPSR() (changeLogs []ChangeLog) {
	for parentID, v := range p.AlleNutzerImFremdsystem {
		err := p.itsl.CreateParentChildRelaionship(parentID, p.ElternKindBeziehungImFremdsystem[parentID])
		fmt.Println(v.Username, ": ", p.ElternKindBeziehungImFremdsystem[parentID])
		if err != nil {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: parentID,
				PSR:         true,
				Error:       err.Error(),
			})
		} else {
			changeLogs = append(changeLogs, ChangeLog{
				UserOrGroup: parentID,
				PSR:         true,
			})
		}
	}
	return changeLogs
}
