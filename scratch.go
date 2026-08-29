package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/company/hrbot/internal/domain/vacancy"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://hrbot:ae5c47c62904f1f561608df2@localhost:5432/hrbot?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo := vacancy.NewRepository(pool)
	
	salaryText := "Kelishiladi"
	loc := "Samarqand"
	curr := "UZS"
	desc := "Test description"
	reqs := "Test reqs"
	bens := "Test bens"
	slug := "test-vacancy-" + time.Now().Format("150405")

	err = repo.Create(ctx, &vacancy.Vacancy{
		Title: "Test Vacancy",
		Slug: slug,
		Location: &loc,
		SalaryText: &salaryText,
		SalaryCurrency: &curr,
		Description: &desc,
		Requirements: &reqs,
		Benefits: &bens,
		Status: "draft",
	})

	if err != nil {
		fmt.Printf("Error creating vacancy: %v\n", err)
	} else {
		fmt.Println("Vacancy created successfully!")
	}
}
