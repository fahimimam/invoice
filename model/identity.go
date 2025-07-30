package model

type Identity struct {
	Id               string `json:"id"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
	Dob              string `json:"dob"`
	PresentAddress   string `json:"presentAddress"`
	PermanentAddress string `json:"permanentAddress"`
	Gender           string `json:"gender"`
	NationalID       string `json:"nationalID"`
	Owner            string `json:"owner"`
}
