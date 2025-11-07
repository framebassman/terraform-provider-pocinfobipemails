terraform {
  required_providers {
    pocinfobipemails = {
      source = "hashicorp.com/edu/pocinfobipemails"
    }
  }
}

variable "infobip_base_url" {}
variable "infobip_api_key" {}

provider "pocinfobipemails" {
  base_url = var.infobip_base_url
  api_key  = var.infobip_api_key
}

data "pocinfobipemails_email_templates" "edu" {}

output "edu_email_templates" {
  value = data.pocinfobipemails_email_templates.edu
}