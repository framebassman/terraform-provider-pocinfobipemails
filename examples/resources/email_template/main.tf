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

resource "pocinfobipemails_email_template" "edu" {
  name         = "Welcome email take3"
  from         = "Romashov <noreply@romashov.tech>"
  reply_to     = "support@example.com"
  subject      = "Welcome to Infobip1"
  preheader    = "Welcome to Infobip"
  html         = "<html><head></head><body><h2>Welcome to Infobip</h2></body></html>"
  landing_page = "1_2345"
}
