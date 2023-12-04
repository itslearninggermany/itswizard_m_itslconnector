package itswizard_m_itslconnector

import (
	"fmt"
	itswizard_basic "github.com/itslearninggermany/itswizard_m_basic"
	imses "github.com/itslearninggermany/itswizard_m_imses"
)

func (p *Itslconnector) holeAlleNutzerAusItslearning() (map[string]itswizard_basic.Person, error) {
	itslUsers := make(map[string]itswizard_basic.Person)
	out := p.itsl.ReadPersonsForGroup(p.rootGroupIdInItslearning, p.organisationID, p.instiutionID)
	if out.Err != nil {
		return nil, out.Err
	}
	for i := 0; i < len(out.Persons); i++ {
		if p.profilesToSync[out.Persons[i].Profile] {
			itslUsers[out.Persons[i].PersonSyncKey] = out.Persons[i]
		}
	}
	p.alleNutzerInItslearning = itslUsers
	return itslUsers, nil
}

func (p *Itslconnector) hohleAlleMembershipAusItslearning() (map[string][]imses.Membership, error) {
	groupMembers := make(map[string][]imses.Membership)
	for i := 0; i < len(p.Gruppen); i++ {
		if p.Gruppen[i].ToSync {
			out, err, resp := p.itsl.ReadMembershipsForGroup(p.Gruppen[i].GroupSyncKey)
			if err != nil {
				fmt.Println(resp)
				continue
				//				return nil, errors.New(resp)
			}
			tmp := []imses.Membership{}
			for su := 0; su < len(out); su++ {
				if out[su].Profile != "Administrator" {
					tmp = append(tmp, out[su])
				}
			}
			groupMembers[p.Gruppen[i].GroupSyncKey] = tmp
		}
	}
	return groupMembers, nil
}

func (p *Itslconnector) GetItslearningUsers() map[string]itswizard_basic.Person {
	return p.alleNutzerInItslearning
}

func (p *Itslconnector) GetItslearningMemberships() map[string][]imses.Membership {
	return p.mitgliedschaftenInItslearning
}

func (p *Itslconnector) GetElternKindBeziehungInItslearning() map[string][]string {
	out, _ := p.holeElternKindBeziehungAusItslearning()
	return out
}

func (p *Itslconnector) holeElternKindBeziehungAusItslearning() (map[string][]string, error) {
	mb, err := p.itsl.ReadParenChildRelationship(p.rootGroupIdInItslearning, p.organisationID, p.instiutionID)
	if err != nil {
		return nil, err
	}
	p.elternKindBeziehungInItslearning = mb
	return mb, nil
}
