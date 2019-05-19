package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const serviceKeyResource = `

data "cloudfoundry_org" "org" {
  name = "%s"
}
data "cloudfoundry_space" "space" {
  name = "%s"
  org = "${data.cloudfoundry_org.org.id}"
}
data "cloudfoundry_service" "test-service" {
  name = "%s"
}

resource "cloudfoundry_service_instance" "test-service-instance" {
	name = "test-service-instance"
  space = "${data.cloudfoundry_space.space.id}"
  service_plan = "${data.cloudfoundry_service.test-service.service_plans["%s"]}"
}

resource "cloudfoundry_service_key" "test-service-instance-key" {
	name = "test-service-instance-key"
	service_instance = "${cloudfoundry_service_instance.test-service-instance.id}"

	params {
		"key1" = "aaaa"
		"key2" = "bbbb"
	}
}
`

func TestAccServiceKey_normal(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)
	serviceName1, _, servicePlan := getTestServiceBrokers(t)

	ref := "cloudfoundry_service_key.test-service-instance-key"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			CheckDestroy: testAccCheckServiceKeyDestroyed(
				"test-service-instance-key", ref),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(serviceKeyResource,
						orgName, spaceName,
						serviceName1, servicePlan),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceKeyExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "test-service-instance-key"),
					),
				},
			},
		})
}

func testAccCheckServiceKeyExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service instance '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes
		serviceKey, _, err := session.ClientV2.GetServiceKey(id)
		if err != nil {
			return err
		}

		if err = assertEquals(attributes, "name", serviceKey.Name); err != nil {
			return err
		}

		// normalized := normalizeMap(serviceKey.Credentials, make(map[string]interface{}), "", "_")
		return assertMapEquals("credentials", attributes, serviceKey.Credentials)
	}
}

func testAccCheckServiceKeyDestroyed(name, serviceInstanceKey string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[serviceInstanceKey]
		if !ok {
			return fmt.Errorf("service instance '%s' not found in terraform state", serviceInstanceKey)
		}

		_, _, err := session.ClientV2.GetServiceKey(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("service instance with name '%s' still exists in cloud foundry", name)
		}
		return nil
	}
}
