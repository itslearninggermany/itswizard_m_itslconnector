package itswizard_m_itslconnector

/*
import (
	"github.com/itslearninggermany/blusd"
)

type ItslConnectorForLusd struct {
	RootGroupIdInItslearning string
	Itsl                     *imses.Request
	OrganisationID           uint
	InstitutionID            uint
	blusd 			*blusd.BlusdConnection
}

func NewItslConnectorForLengde(input ItslConnectorForLusd) (*Itslconnector, error) {
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
	itslc.Gruppen = nil
	membitsl, err := itslc.hohleAlleMembershipAusItslearning()
	if err != nil {

		return nil, err
	}
	itslc.mitgliedschaftenInItslearning = membitsl
	itslc.mitgliedschaftenImFremdsytem = nil
	itslc.alleNutzerImFremdsystem = nil
	return itslc, nil
}
*/
