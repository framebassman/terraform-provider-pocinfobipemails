terraform {
  required_providers {
    poc_infobip_emails = {
      source = "hashicorp.com/edu/poc-infobip-emails"
    }
  }
}

variable "infobip_base_url" {}
variable "infobip_api_key" {}

provider "poc_infobip_emails" {
  base_url = "${var.infobip_base_url}"
  api_key = "${var.infobip_api_key}"
}

data "poc_infobip_emails_email_templates" "edu" {}

output "edu_email_templates" {
  value = data.poc_infobip_emails_email_templates.edu
}
