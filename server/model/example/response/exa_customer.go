package response

import "github.com/isgvto/gin-vue-admin-gblog/server/model/example"

type ExaCustomerResponse struct {
	Customer example.ExaCustomer `json:"customer"`
}
