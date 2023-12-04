package itswizard_m_itslconnector

import (
	itswizard_basic "github.com/itslearninggermany/itswizard_m_basic"
	imses "github.com/itslearninggermany/itswizard_m_imses"
)

/*
Alle Nutzer in Itslearning
Alle Gruppen in itslearning [groupID][]itswizardPerson
Alle Nutzer im anderen System [string]itswizard_basicPerson
Alle Gruppen im anderen System [groupID][]itswizardPerson
*/

type Itslconnector struct {
	Fremdsystem                      string
	Gruppen                          []itswizard_basic.Group
	itsl                             *imses.Request
	profilesToSync                   map[string]bool
	rootGroupIdInItslearning         string
	organisationID                   uint
	instiutionID                     uint
	alleGruppenAusAAD                []itswizard_basic.Group
	alleNutzerInItslearning          map[string]itswizard_basic.Person
	mitgliedschaftenInItslearning    map[string][]imses.Membership
	elternKindBeziehungInItslearning map[string][]string
	alleNutzerImFremdsystem          map[string]itswizard_basic.Person
	mitgliedschaftenImFremdsytem     map[string][]imses.Membership
	elternKindBeziehungImFremdsystem map[string][]string //  (( elternId --> []SchülerIds))
}

/*
Variablen die gebracht werden
p.itsl
p.rootGroupIdInItslearsning
p.organisationID
p.instiutionID
p.profilesToSync
*/

func (p *Itslconnector) GetUsersFromSourceSystem() map[string]itswizard_basic.Person {
	return p.alleNutzerImFremdsystem
}

func (p *Itslconnector) GetMembershipsFromSourceSystem() map[string][]imses.Membership {
	return p.mitgliedschaftenImFremdsytem
}
