package AutomationService

import "fmt"

func (s *AutomationService) TestAutomation(value string) {
	fmt.Println("---- TEST START ----")
	fmt.Println(value)
	fmt.Println("---- TEST END ----")
}
