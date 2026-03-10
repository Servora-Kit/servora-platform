variable "VERSION" {
  default = "dev"
}

target "_common" {
  context    = "."
  dockerfile = "Dockerfile"
  output = ["type=docker"]
  args = {
    VERSION = VERSION
  }
}

target "servora" {
  inherits = ["_common"]
  args = {
    SERVICE_NAME = "servora"
  }
  tags = [
    "servora/servora-service:${VERSION}",
    "servora/servora-service:latest",
  ]
}

target "sayhello" {
  inherits = ["_common"]
  args = {
    SERVICE_NAME = "sayhello"
  }
  tags = [
    "servora/sayhello-service:${VERSION}",
    "servora/sayhello-service:latest",
  ]
}

group "default" {
  targets = ["servora", "sayhello"]
}
