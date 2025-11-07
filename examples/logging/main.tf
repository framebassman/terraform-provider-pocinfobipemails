terraform {
  required_providers {
    pocinfobipemails = {
      source = "hashicorp.com/edu/pocinfobipemails"
    }
  }
}

variable "infobip_api_key" {}

provider "pocinfobipemails" {
  base_url = "https://51dg6z.api.infobip.com"
  api_key  = var.infobip_api_key
}

data "pocinfobipemails_email_templates" "edu" {}
