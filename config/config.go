package config

import (
    "log"
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type Config struct {
    DBUser      string
    DBPass      string
    DBHost      string
    DBPort      string
    DBName      string
    JWTSecret   string
    JWTExpiresH int
    Port        string
}

func Load() *Config {
    _ = godotenv.Load()

    exp, _ := strconv.Atoi(os.Getenv("JWT_EXPIRES_H"))
    if exp == 0 {
        exp = 24 // default 24 jam
    }

    cfg := &Config{
        DBUser:      os.Getenv("DB_USER"),
        DBPass:      os.Getenv("DB_PASS"),
        DBHost:      os.Getenv("DB_HOST"),
        DBPort:      os.Getenv("DB_PORT"),
        DBName:      os.Getenv("DB_NAME"),
        JWTSecret:   os.Getenv("JWT_SECRET"),
        JWTExpiresH: exp,
        Port:        os.Getenv("PORT"),
    }

    if cfg.Port == "" {
        cfg.Port = "8080"
    }

    if cfg.DBHost == "" {
        log.Fatal("Missing DB config in .env")
    }

    return cfg
}
