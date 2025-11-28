terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~>7.9.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~>6.22.0"
    }
  }
}
