package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudHsm2Hsm_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudHsm2HsmDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSCloudHsm2Hsm(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudHsm2HsmExists("aws_cloudhsm_v2_hsm.hsm"),
					resource.TestCheckResourceAttrSet("aws_cloudhsm_v2_hsm.hsm", "hsm_id"),
					resource.TestCheckResourceAttrSet("aws_cloudhsm_v2_hsm.hsm", "hsm_state"),
					resource.TestCheckResourceAttrSet("aws_cloudhsm_v2_hsm.hsm", "hsm_eni_id"),
					resource.TestCheckResourceAttrSet("aws_cloudhsm_v2_hsm.hsm", "ip_address"),
				),
			},
		},
	})
}

func testAccAWSCloudHsm2Hsm() string {
	return fmt.Sprintf(`
variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "cloudhsm2_test_vpc" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "terraform-testacc-aws_cloudhsm_v2_hsm-resource-basic"
  }
}

resource "aws_subnet" "cloudhsm2_test_subnets" {
  count                   = 2
  vpc_id                  = "${aws_vpc.cloudhsm2_test_vpc.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = false
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags {
    Name = "tf-acc-aws_cloudhsm_v2_hsm-resource-basic"
  }
}

resource "aws_cloudhsm_v2_cluster" "cloudhsm_v2_cluster" {
  hsm_type = "hsm1.medium"  
  subnet_ids = ["${aws_subnet.cloudhsm2_test_subnets.*.id}"]
  tags {
    Name = "tf-acc-aws_cloudhsm_v2_hsm-resource-basic-%d"
  }
}

resource "aws_cloudhsm_v2_hsm" "hsm" {
  subnet_id = "${aws_subnet.cloudhsm2_test_subnets.0.id}"
  cluster_id = "${aws_cloudhsm_v2_cluster.cloudhsm_v2_cluster.cluster_id}"
}
`, acctest.RandInt())
}

func testAccCheckAWSCloudHsm2HsmDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudhsmv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudhsm_v2_hsm" {
			continue
		}

		var hsm *cloudhsmv2.Hsm
		out, err := conn.DescribeClusters(&cloudhsmv2.DescribeClustersInput{})
		for _, c := range out.Clusters {
			for _, h := range c.Hsms {
				if aws.StringValue(h.HsmId) == rs.Primary.ID {
					hsm = h
				}
			}
		}

		if err != nil {
			return err
		}

		if hsm != nil && aws.StringValue(hsm.State) != "DELETED" {
			return fmt.Errorf("HSM still exists:\n%s", hsm)
		}
	}

	return nil
}

func testAccCheckAWSCloudHsm2HsmExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}
