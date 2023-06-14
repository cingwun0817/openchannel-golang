package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
)

var day int = 2

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
		for d := 0; d < day; d++ {
			date := today.AddDate(0, 0, 0-d)

			eventStat := handleEvent(ctx, session, storeId, date.Format("2006-01-02"))
			otsStat := handleOts(ctx, session, storeId, date.Format("2006-01-02"))
			playStat := handlePlay(ctx, session, storeId, date.Format("2006-01-02"))

			peopleStat := merge(storeId, date.Format("2006-01-02"), eventStat, otsStat, playStat)

			batchInsert(ctx, session, peopleStat)

			fmt.Printf("store_id: %s \t date: %s \t total: %d\n", storeId, date.Format("2006-01-02"), len(peopleStat))
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

type EventStat struct {
	CountAttention    int
	CountMaleChild    int
	CountMaleYoung    int
	CountMaleAdult    int
	CountMaleSenior   int
	CountFemaleChild  int
	CountFemaleYoung  int
	CountFemaleAdult  int
	CountFemaleSenior int
	CountUnknown      int
	Attention         float64
}

func handleEvent(ctx context.Context, session *gocql.Session, pStoreId string, pDate string) map[string]*EventStat {
	stat := make(map[string]*EventStat)

	scanner := session.Query("SELECT store_id, date, hour, media_id, age, attention, gender FROM oc.quividi_event WHERE store_id = ? AND date = ?", pStoreId, pDate).WithContext(ctx).Iter().Scanner()
	for scanner.Next() {
		var store_id, media_id string
		var age, gender int
		var date string
		var hour float32
		var attention float64

		err := scanner.Scan(&store_id, &date, &hour, &media_id, &age, &attention, &gender)
		if err != nil {
			log.Fatal(err)
		}

		key := fmt.Sprintf("%s.%s.%g.%s", store_id, date, hour, media_id)
		_, exist := stat[key]
		if !exist {
			stat[key] = &EventStat{}
		}

		if gender == 0 || age == 0 {
			stat[key].CountUnknown++
		} else {
			if gender == 1 {
				switch age {
				case 1:
					stat[key].CountMaleChild++
				case 2:
					stat[key].CountMaleYoung++
				case 3:
					stat[key].CountMaleAdult++
				case 4:
					stat[key].CountMaleSenior++
				}
			} else if gender == 2 {
				switch age {
				case 1:
					stat[key].CountFemaleChild++
				case 2:
					stat[key].CountFemaleYoung++
				case 3:
					stat[key].CountFemaleAdult++
				case 4:
					stat[key].CountFemaleSenior++
				}
			}

			// only not unknown target can compute attention (date: 20230606 <leo.wang>)
			stat[key].Attention += attention
			stat[key].CountAttention++
		}
	}

	return stat
}

type OtsStat struct {
	CountView int
}

func handleOts(ctx context.Context, session *gocql.Session, pStoreId string, pDate string) map[string]*OtsStat {
	stat := make(map[string]*OtsStat)

	scanner := session.Query("SELECT store_id, date, hour, media_id, count FROM oc.quividi_ots WHERE store_id = ? AND date = ?", pStoreId, pDate).WithContext(ctx).Iter().Scanner()
	for scanner.Next() {
		var store_id, media_id string
		var count int
		var date string
		var hour float32

		err := scanner.Scan(&store_id, &date, &hour, &media_id, &count)
		if err != nil {
			log.Fatal(err)
		}

		key := fmt.Sprintf("%s.%s.%g.%s", store_id, date, hour, media_id)
		_, exist := stat[key]
		if !exist {
			stat[key] = &OtsStat{}
		}

		stat[key].CountView += count
	}

	return stat
}

type PlayStat struct {
	CountView int
	StoreId   string
	Date      string
	Hour      float32
	MediaId   string
}

func handlePlay(ctx context.Context, session *gocql.Session, pStoreId string, pDate string) map[string]*PlayStat {
	stat := make(map[string]*PlayStat)

	scanner := session.Query("SELECT store_id, date, hour, media_id, count FROM oc.quividi_playcnt WHERE store_id = ? AND date = ?", pStoreId, pDate).WithContext(ctx).Iter().Scanner()
	for scanner.Next() {
		var store_id, media_id string
		var count int
		var date string
		var hour float32

		err := scanner.Scan(&store_id, &date, &hour, &media_id, &count)
		if err != nil {
			log.Fatal(err)
		}

		key := fmt.Sprintf("%s.%s.%g.%s", store_id, date, hour, media_id)
		_, exist := stat[key]
		if !exist {
			stat[key] = &PlayStat{}
			stat[key].StoreId = store_id
			stat[key].Date = date
			stat[key].Hour = hour
			stat[key].MediaId = media_id
		}

		stat[key].CountView += count
	}

	return stat
}

type PeopleStat struct {
	StoreId           string
	Date              string
	Hour              float32
	MediaId           string
	CountPlay         int
	CountPeople       int
	Impression        int
	Attention         float64
	CountMale         int
	CountMaleChild    int
	CountMaleYoung    int
	CountMaleAdult    int
	CountMaleSenior   int
	CountFemale       int
	CountFemaleChild  int
	CountFemaleYoung  int
	CountFemaleAdult  int
	CountFemaleSenior int
	CountUnknown      int
}

func merge(pStoreId string, pDate string, eventStat map[string]*EventStat, otsStat map[string]*OtsStat, play map[string]*PlayStat) []PeopleStat {
	stat := []PeopleStat{}

	for k, v := range play {

		rowStat := PeopleStat{}
		rowStat.StoreId = v.StoreId
		rowStat.Date = v.Date
		rowStat.Hour = v.Hour
		rowStat.MediaId = v.MediaId
		rowStat.CountPlay = v.CountView

		_, existEvent := eventStat[k]
		if existEvent {
			rowStat.Impression = eventStat[k].CountAttention
			rowStat.Attention = eventStat[k].Attention
			rowStat.CountMale = eventStat[k].CountMaleChild + eventStat[k].CountMaleYoung + eventStat[k].CountMaleAdult + eventStat[k].CountMaleSenior
			rowStat.CountMaleChild = eventStat[k].CountMaleChild
			rowStat.CountMaleYoung = eventStat[k].CountMaleYoung
			rowStat.CountMaleAdult = eventStat[k].CountMaleAdult
			rowStat.CountMaleSenior = eventStat[k].CountMaleSenior
			rowStat.CountFemale = eventStat[k].CountFemaleChild + eventStat[k].CountFemaleYoung + eventStat[k].CountFemaleAdult + eventStat[k].CountFemaleSenior
			rowStat.CountFemaleChild = eventStat[k].CountFemaleChild
			rowStat.CountFemaleYoung = eventStat[k].CountFemaleYoung
			rowStat.CountFemaleAdult = eventStat[k].CountFemaleAdult
			rowStat.CountFemaleSenior = eventStat[k].CountFemaleSenior
			rowStat.CountUnknown = eventStat[k].CountUnknown
		}

		_, existOts := otsStat[k]
		if existOts {
			rowStat.CountPeople = otsStat[k].CountView
		}

		stat = append(stat, rowStat)
	}

	return stat
}

func batchInsert(ctx context.Context, session *gocql.Session, peopleStat []PeopleStat) {
	batch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)

	for _, stat := range peopleStat {
		batch.Entries = append(batch.Entries, gocql.BatchEntry{
			Stmt:       "INSERT INTO oc.quividi_people_hour_analyze (date, hour, store_id, media_id, play_count, people_count, impression, male, male_child, male_young, male_adult, male_senior, female, female_child, female_young, female_adult, female_senior, unknown, attention) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			Args:       []interface{}{stat.Date, stat.Hour, stat.StoreId, stat.MediaId, stat.CountPlay, stat.CountPeople, stat.Impression, stat.CountMale, stat.CountMaleChild, stat.CountMaleYoung, stat.CountMaleAdult, stat.CountMaleSenior, stat.CountFemale, stat.CountFemaleChild, stat.CountFemaleYoung, stat.CountFemaleAdult, stat.CountFemaleSenior, stat.CountUnknown, stat.Attention},
			Idempotent: true,
		})
	}

	err := session.ExecuteBatch(batch)
	if err != nil {
		log.Fatal(err)
	}
}
