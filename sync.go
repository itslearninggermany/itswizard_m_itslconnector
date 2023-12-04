package itswizard_m_itslconnector

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	itswizard_basic "github.com/itslearninggermany/itswizard_m_basic"
	imses "github.com/itslearninggermany/itswizard_m_imses"
)

type UserUpdate struct {
	Update   string
	UserOld  itswizard_basic.Person
	UsersNew itswizard_basic.Person
}

func (p *Itslconnector) GetUserToDelete() map[string]itswizard_basic.Person {
	itslUsers := mapset.NewSet()
	for k, _ := range p.alleNutzerInItslearning {
		itslUsers.Add(k)
	}
	sourceUsers := mapset.NewSet()
	for k, _ := range p.alleNutzerImFremdsystem {
		sourceUsers.Add(k)
	}

	set := itslUsers.Intersect(sourceUsers)
	set = itslUsers.Difference(set)

	deleteUsers := make(map[string]itswizard_basic.Person)
	for {
		erg := set.Pop()
		if erg == nil {
			break
		}
		deleteUsers[fmt.Sprint(erg)] = p.alleNutzerInItslearning[fmt.Sprint(erg)]
	}
	return deleteUsers
}

func (p *Itslconnector) GetUserToImport() map[string]itswizard_basic.Person {
	itslUsers := mapset.NewSet()
	for k, _ := range p.alleNutzerInItslearning {
		itslUsers.Add(k)
	}
	sourceUsers := mapset.NewSet()
	for k, _ := range p.alleNutzerImFremdsystem {
		sourceUsers.Add(k)
	}

	set := itslUsers.Intersect(sourceUsers)
	set = sourceUsers.Difference(set)

	importUser := make(map[string]itswizard_basic.Person)
	for {
		erg := set.Pop()
		if erg == nil {
			break
		}
		importUser[erg.(string)] = p.alleNutzerImFremdsystem[erg.(string)]
	}
	return importUser
}

func (p *Itslconnector) GetUsersToUpdate() map[string]UserUpdate {
	update := make(map[string]UserUpdate)
	itslUsers := mapset.NewSet()
	for k, _ := range p.alleNutzerInItslearning {
		itslUsers.Add(k)
	}
	sourceUsers := mapset.NewSet()
	for k, _ := range p.alleNutzerImFremdsystem {
		sourceUsers.Add(k)
	}

	set := itslUsers.Intersect(sourceUsers)

	for {
		personToCheck := set.Pop()
		if personToCheck == nil {
			break
		}
		tmpString := ""
		toUpdate := false
		if p.alleNutzerImFremdsystem[personToCheck.(string)].FirstName != p.alleNutzerInItslearning[personToCheck.(string)].FirstName {
			tmpString = tmpString + "Vorname wird aktualisiert. "
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].LastName != p.alleNutzerInItslearning[personToCheck.(string)].LastName {
			tmpString = tmpString + "Nachname wird aktualisiert. "
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].Username != p.alleNutzerInItslearning[personToCheck.(string)].Username {
			tmpString = tmpString + "Nutzername wird aktualisiert."
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].Profile != p.alleNutzerInItslearning[personToCheck.(string)].Profile {
			tmpString = tmpString + "Profil wird aktualisiert. "
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].Email != p.alleNutzerInItslearning[personToCheck.(string)].Email {
			tmpString = tmpString + "Email. wird aktualisiert"
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].Phone != p.alleNutzerInItslearning[personToCheck.(string)].Phone {
			tmpString = tmpString + "Tel. wird aktualisiert"
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].Mobile != p.alleNutzerInItslearning[personToCheck.(string)].Mobile {
			tmpString = tmpString + "Mobile wird aktualisiert."
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].Street1 != p.alleNutzerInItslearning[personToCheck.(string)].Street1 {
			tmpString = tmpString + "Adresse(1) wird aktualisiert."
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].Street2 != p.alleNutzerInItslearning[personToCheck.(string)].Street2 {
			tmpString = tmpString + "Adresse(2) wird aktualisiert."
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].Postcode != p.alleNutzerInItslearning[personToCheck.(string)].Postcode {
			tmpString = tmpString + "Postleitzahl wird aktualisiert."
			toUpdate = true
		}
		if p.alleNutzerImFremdsystem[personToCheck.(string)].City != p.alleNutzerInItslearning[personToCheck.(string)].City {
			tmpString = tmpString + "Stadt wird aktualisiert."
			toUpdate = true
		}

		if toUpdate {
			tt := UserUpdate{
				Update:   tmpString,
				UsersNew: p.alleNutzerImFremdsystem[personToCheck.(string)],
				UserOld:  p.alleNutzerInItslearning[personToCheck.(string)],
			}
			update[personToCheck.(string)] = tt
		}
	}
	return update
}

func (p *Itslconnector) GetMembershipToDelete() map[string][]imses.Membership {
	GroupMembershipsToDelete := make(map[string][]imses.Membership)
	for k, membershipsInSource := range p.mitgliedschaftenImFremdsytem {
		membershipInItsl := p.mitgliedschaftenInItslearning[k]

		tmpMembersItsl := mapset.NewSet()
		for i := 0; i < len(membershipInItsl); i++ {
			tmpMembersItsl.Add(membershipInItsl[i].PersonID)
		}
		tmpMembersSource := mapset.NewSet()
		for i := 0; i < len(membershipsInSource); i++ {
			tmpMembersSource.Add(membershipsInSource[i].PersonID)
		}
		set := tmpMembersItsl.Intersect(tmpMembersSource)
		set = tmpMembersItsl.Difference(set)

		allPersonToDelete := []string{}
		for {
			erg := set.Pop()
			if erg == nil {
				break
			}
			allPersonToDelete = append(allPersonToDelete, erg.(string))
		}
		tmppp := []imses.Membership{}
		for i := 0; i < len(allPersonToDelete); i++ {
			for g := 0; g < len(membershipInItsl); g++ {
				if allPersonToDelete[i] == membershipInItsl[g].PersonID {
					tmppp = append(tmppp, membershipInItsl[g])
					break
				}
			}
		}
		GroupMembershipsToDelete[k] = tmppp
	}
	return GroupMembershipsToDelete
}

func (p *Itslconnector) GetMembershipToImport() map[string][]imses.Membership {
	GroupMembershipsToImport := make(map[string][]imses.Membership)
	for k, membershipsInSource := range p.mitgliedschaftenImFremdsytem {
		membershipInItsl := p.mitgliedschaftenInItslearning[k]

		tmpMembersItsl := mapset.NewSet()
		for i := 0; i < len(membershipInItsl); i++ {
			tmpMembersItsl.Add(membershipInItsl[i].PersonID)
		}
		tmpMembersSource := mapset.NewSet()
		for i := 0; i < len(membershipsInSource); i++ {
			tmpMembersSource.Add(membershipsInSource[i].PersonID)
		}
		set := tmpMembersItsl.Intersect(tmpMembersSource)
		set = tmpMembersSource.Difference(set)

		allPersonToImport := []string{}
		for {
			erg := set.Pop()
			if erg == nil {
				break
			}
			allPersonToImport = append(allPersonToImport, erg.(string))
		}
		tmppp := []imses.Membership{}
		for i := 0; i < len(allPersonToImport); i++ {
			for g := 0; g < len(membershipsInSource); g++ {
				if allPersonToImport[i] == membershipsInSource[g].PersonID {
					tmppp = append(tmppp, membershipsInSource[g])
					break
				}
			}
		}
		GroupMembershipsToImport[k] = tmppp
	}
	return GroupMembershipsToImport
}

func (p *Itslconnector) ParentStudentUpdate() {

}
