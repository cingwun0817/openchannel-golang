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

			// insert to store_analyze
			taStore := getStoreTargetAudience(ctx, session, storeId, startDate.Format("2006-01-02"), today.Format("2006-01-02"))

			if storeId != taStore.StoreId {
				continue
			}

			err := session.Query(
				"INSERT INTO oc.ta_store_analyze (store_id, day, female, male, child, young, adult, senior, female_child, female_young, female_adult, female_senior, male_child, male_young, male_adult, male_senior, play_count, people_count, impression) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				taStore.StoreId,
				day,
				taStore.Female,
				taStore.Male,
				taStore.Child,
				taStore.Young,
				taStore.Adult,
				taStore.Senior,
				taStore.FemaleChild,
				taStore.FemaleYoung,
				taStore.FemaleAdult,
				taStore.FemaleSenior,
				taStore.MaleChild,
				taStore.MaleYoung,
				taStore.MaleAdult,
				taStore.MaleSenior,
				taStore.PlayCount,
				taStore.PeopleCount,
				taStore.Impression,
			).WithContext(ctx).Exec()
			if err != nil {
				log.Fatal(err)
			}

			// insert to store_media_analyze
			taStoreMedias := getStoreMediaTargetAudience(ctx, session, storeId, startDate.Format("2006-01-02"), today.Format("2006-01-02"))

			for _, taStoreMedia := range taStoreMedias {
				if storeId != taStoreMedia.StoreId {
					continue
				}
				err := session.Query(
					"INSERT INTO oc.ta_store_media_analyze (store_id, media_id, day, female, male, child, young, adult, senior, female_child, female_young, female_adult, female_senior, male_child, male_young, male_adult, male_senior, play_count, people_count, impression) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
					taStoreMedia.StoreId,
					taStoreMedia.MediaId,
					day,
					taStoreMedia.Female,
					taStoreMedia.Male,
					taStoreMedia.Child,
					taStoreMedia.Young,
					taStoreMedia.Adult,
					taStoreMedia.Senior,
					taStoreMedia.FemaleChild,
					taStoreMedia.FemaleYoung,
					taStoreMedia.FemaleAdult,
					taStoreMedia.FemaleSenior,
					taStoreMedia.MaleChild,
					taStoreMedia.MaleYoung,
					taStoreMedia.MaleAdult,
					taStoreMedia.MaleSenior,
					taStoreMedia.PlayCount,
					taStoreMedia.PeopleCount,
					taStoreMedia.Impression,
				).WithContext(ctx).Exec()
				if err != nil {
					log.Fatal(err)
				}
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
	StoreId      string
	MediaId      string
	Female       int
	Male         int
	Child        int
	Young        int
	Adult        int
	Senior       int
	FemaleChild  int
	FemaleYoung  int
	FemaleAdult  int
	FemaleSenior int
	MaleChild    int
	MaleYoung    int
	MaleAdult    int
	MaleSenior   int
	PlayCount    int
	PeopleCount  int
	Impression   int
}

func getStoreTargetAudience(ctx context.Context, session *gocql.Session, storeId string, sDate string, eDate string) Target {
	var store_id string
	var female, female_child, female_young, female_adult, female_senior, male, male_child, male_young, male_adult, male_senior, play_count, people_count, impression int

	err := session.Query(
		"SELECT store_id, SUM(female) as female, SUM(female_child) as female_child, SUM(female_young) as female_young, SUM(female_adult) as female_adult, SUM(female_senior) as female_senior, SUM(male) as male, SUM(male_child) as male_child, SUM(male_young) as male_young, SUM(male_adult) as male_adult, SUM(male_senior) as male_senior, SUM(play_count) as play_count, SUM(people_count) as people_count, SUM(impression) as impression FROM oc.quividi_people_hour_analyze_by_store_date_hour_media WHERE store_id = ? AND date >= ? AND date <= ?",
		storeId,
		sDate,
		eDate,
	).WithContext(ctx).Scan(&store_id, &female, &female_child, &female_young, &female_adult, &female_senior, &male, &male_child, &male_young, &male_adult, &male_senior, &play_count, &people_count, &impression)
	if err != nil {
		log.Fatal(err)
	}

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

		target.PlayCount = play_count
		target.PeopleCount = people_count
		target.Impression = impression
	}

	return target
}

func getStoreMediaTargetAudience(ctx context.Context, session *gocql.Session, storeId string, sDate string, eDate string) map[string]Target {
	targets := map[string]Target{}

	scanner := session.Query(
		"SELECT store_id, media_id, female, female_child, female_young, female_adult, female_senior, male, male_child, male_young, male_adult, male_senior, play_count, people_count, impression FROM oc.quividi_people_hour_analyze_by_store_date_hour_media WHERE store_id = ? AND date >= ? AND date <= ?",
		storeId,
		sDate,
		eDate,
	).WithContext(ctx).Iter().Scanner()

	for scanner.Next() {
		var store_id, media_id string
		var female, female_child, female_young, female_adult, female_senior, male, male_child, male_young, male_adult, male_senior, play_count, people_count, impression int

		err := scanner.Scan(&store_id, &media_id, &female, &female_child, &female_young, &female_adult, &female_senior, &male, &male_child, &male_young, &male_adult, &male_senior, &play_count, &people_count, &impression)
		if err != nil {
			log.Fatal(err)
		}

		var target Target
		if _, ok := targets[media_id]; ok {
			target = targets[media_id]

			if female != 0 && male != 0 {
				target.Female += female
				target.Male += male
				target.Child += female_child + male_child
				target.Young += female_young + male_young
				target.Adult += female_adult + male_adult
				target.Senior += female_senior + male_senior
				target.FemaleChild += female_child
				target.FemaleYoung += female_young
				target.FemaleAdult += female_adult
				target.FemaleSenior += female_senior
				target.MaleChild += male_child
				target.MaleYoung += male_young
				target.MaleAdult += male_adult
				target.MaleSenior += male_senior

				target.PlayCount += play_count
				target.PeopleCount += people_count
				target.Impression += impression
			}
		} else {
			target = Target{}
			target.StoreId = store_id
			target.MediaId = media_id
		}

		targets[media_id] = target
	}

	return targets
}
