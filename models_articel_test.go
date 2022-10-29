package main

import "testing"

// Test the function that fetches all articles
func TestGetAllArticles(t *testing.T) {
	alist, err := getAllTraefikRouters()
	if err != nil {
		t.Fail()
	}
	// Check that the length of the list of articles returned is the
	// same as the length of the global variable holding the list
	if len(alist) != len(TraefikRouterList) {
		t.Fail()
	}

	// Check that each member is identical
	for i, v := range alist {
		if v.Name != TraefikRouterList[i].Name ||
			v.Status != TraefikRouterList[i].Status ||
			v.Rule != TraefikRouterList[i].Rule {

			t.Fail()
			break
		}
	}
}
