env "local" {
  src = "file://schema.sql"
  url = "postgres://leagueofren:localdev123@localhost:5432/leagueofren?sslmode=disable"
  dev = "docker://postgres/16"
}

env "production" {
  src = "file://schema.sql"
  url = env("DATABASE_URL")
  dev = "docker://postgres/16"
}
