package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
)

var days = [5]int{30, 60, 90, 180, 360}

func main() {
	cluster := gocql.NewCluster("172.16.51.118", "172.16.51.120", "172.16.51.121")

	cluster.Keyspace = "oc"
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	cluster.Timeout = 10 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	ctx := context.Background()

	today := time.Now()
	fmt.Printf("Run time: %s\n", today.Format("2006-01-02 15:04:05"))

	storeIds := getStoreIds(ctx, session)
	for _, storeId := range storeIds {
		for _, day := range days {
			startDate := today.AddDate(0, 0, 0-day)

			ta := getStoreTargetAudience(ctx, session, storeId, startDate.Format("2006-01-02"), today.Format("2006-01-02"))

			if storeId != ta.StoreId {
				continue
			}

			err := session.Query(
				"INSERT INTO oc.ta_store_analyze (store_id, day, female, male, child, young, adult, senior, female_child, female_young, female_adult, female_senior, male_child, male_young, male_adult, male_senior, female_rate, male_rate, child_rate, young_rate, adult_rate, senior_rate, female_child_rate, female_young_rate, female_adult_rate, female_senior_rate, male_child_rate, male_young_rate, male_adult_rate, male_senior_rate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				ta.StoreId,
				day,
				ta.Female,
				ta.Male,
				ta.Child,
				ta.Young,
				ta.Adult,
				ta.Senior,
				ta.FemaleChild,
				ta.FemaleYoung,
				ta.FemaleAdult,
				ta.FemaleSenior,
				ta.MaleChild,
				ta.MaleYoung,
				ta.MaleAdult,
				ta.MaleSenior,
				ta.FemaleRate,
				ta.MaleRate,
				ta.ChildRate,
				ta.YoungRate,
				ta.AdultRate,
				ta.SeniorRate,
				ta.FemaleChildRate,
				ta.FemaleYoungRate,
				ta.FemaleAdultRate,
				ta.FemaleSeniorRate,
				ta.MaleChildRate,
				ta.MaleYoungRate,
				ta.MaleAdultRate,
				ta.MaleSeniorRate,
			).WithContext(ctx).Exec()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("store_id: %s \t day: %d \t date_range: %s - %s\n", storeId, day, startDate.Format("2006-01-02"), today.Format("2006-01-02"))
		}
	}
}

func getStoreIds(ctx context.Context, session *gocql.Session) []string {
	var store_ids []string

	scanner := session.Query("SELECT store_id FROM oc.quividi_finish GROUP BY store_id").WithContext(ctx).Iter().Scanner()
	for scanner.Next() {
		var store_id string
		err := scanner.Scan(&store_id)
		if err != nil {
			log.Fatal(err)
		}

		store_ids = append(store_ids, store_id)
	}

	return store_ids
}

type Target struct {
	StoreId          string
	Female           int
	Male             int
	Child            int
	Young            int
	Adult            int
	Senior           int
	FemaleChild      int
	FemaleYoung      int
	FemaleAdult      int
	FemaleSenior     int
	MaleChild        int
	MaleYoung        int
	MaleAdult        int
	MaleSenior       int
	FemaleRate       float64
	MaleRate         float64
	ChildRate        float64
	YoungRate        float64
	AdultRate        float64
	SeniorRate       float64
	FemaleChildRate  float64
	FemaleYoungRate  float64
	FemaleAdultRate  float64
	FemaleSeniorRate float64
	MaleChildRate    float64
	MaleYoungRate    float64
	MaleAdultRate    float64
	MaleSeniorRate   float64
}

func getStoreTargetAudience(ctx context.Context, session *gocql.Session, storeId string, sDate string, eDate string) Target {
	var store_id string
	var female, female_child, female_young, female_adult, female_senior, male, male_child, male_young, male_adult, male_senior int

	err := session.Query(
		"SELECT store_id, SUM(female) as female, SUM(female_child) as female_child, SUM(female_young) as female_young, SUM(female_adult) as female_adult, SUM(female_senior) as female_senior, SUM(male) as male, SUM(male_child) as male_child, SUM(male_young) as male_young, SUM(male_adult) as male_adult, SUM(male_senior) as male_senior FROM oc.quividi_people_hour_analyze_by_store_date_hour_media WHERE store_id = ? AND date >= ? AND date <= ?",
		storeId,
		sDate,
		eDate,
	).WithContext(ctx).Scan(&store_id, &female, &female_child, &female_young, &female_adult, &female_senior, &male, &male_child, &male_young, &male_adult, &male_senior)
	if err != nil {
		log.Fatal(err)
	}

	total := female + male

	target := Target{}
	target.StoreId = store_id

	if female != 0 && male != 0 {
		target.Female = female
		target.Male = male
		target.Child = female_child + male_child
		target.Young = female_young + male_young
		target.Adult = female_adult + male_adult
		target.Senior = female_senior + male_senior
		target.FemaleChild = female_child
		target.FemaleYoung = female_young
		target.FemaleAdult = female_adult
		target.FemaleSenior = female_senior
		target.MaleChild = male_child
		target.MaleYoung = male_young
		target.MaleAdult = male_adult
		target.MaleSenior = male_senior

		target.FemaleRate = float64(target.Female) / float64(total)
		target.MaleRate = float64(target.Male) / float64(total)
		target.ChildRate = float64(target.Child) / float64(total)
		target.YoungRate = float64(target.Young) / float64(total)
		target.AdultRate = float64(target.Adult) / float64(total)
		target.SeniorRate = float64(target.Senior) / float64(total)
		target.FemaleChildRate = float64(target.FemaleChild) / float64(total)
		target.FemaleYoungRate = float64(target.FemaleYoung) / float64(total)
		target.FemaleAdultRate = float64(target.FemaleAdult) / float64(total)
		target.FemaleSeniorRate = float64(target.FemaleSenior) / float64(total)
		target.MaleChildRate = float64(target.MaleChild) / float64(total)
		target.MaleYoungRate = float64(target.MaleYoung) / float64(total)
		target.MaleAdultRate = float64(target.MaleAdult) / float64(total)
		target.MaleSeniorRate = float64(target.MaleSenior) / float64(total)
	}

	return target
}
