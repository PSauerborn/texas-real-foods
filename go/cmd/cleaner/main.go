package main


import (
    "fmt"
    "time"
    "sync"
    "strconv"
    "context"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/utils"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
            "clear_interval_minutes": "1",
            "retention_period_days": "14",
        },
    )
)

type Persistence struct{
    *utils.BasePostgresPersistence
}

func NewPersistence(url string) *Persistence {
    // create instance of base persistence
    basePersistence := utils.NewPersistence(url)
    return &Persistence{
        basePersistence,
    }
}

func(db *Persistence) ClearData(ts time.Time) error {
    log.Debug(fmt.Sprintf("clearing data from database..."))
    query := `DELETE FROM asset_data_timeseries WHERE event_timestamp < $1`
    _, err := db.Session.Exec(context.Background(), query, ts)
    if err != nil {
        log.Error(fmt.Errorf("unable to delete data from database: %+v", err))
        return err
    }
    return nil
}

func clearData(interval int) error {
    // establish new connection to postgres persistence
    db := NewPersistence(cfg.Get("postgres_url"))
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to connect to postgres server: %+v", err))
        return err
    }
    defer conn.Close()

    ts := time.Now().Add(time.Duration(-interval) * 24 * time.Hour)
    log.Info(fmt.Sprintf("clearing all data before interval %s", ts))
    return db.ClearData(ts)
}

func main() {
    log.SetLevel(log.DebugLevel)
    // convert interval to integer
    interval, err := strconv.Atoi(cfg.Get("clear_interval_minutes"))
    if err != nil {
        panic(fmt.Sprintf("received invalid worker interval %s", cfg.Get("clear_interval_minutes")))
    }
    // convert retention period to integer
    retentionPeriod, err := strconv.Atoi(cfg.Get("retention_period_days"))
    if err != nil {
        panic(fmt.Sprintf("received invalid retention period %s", cfg.Get("clear_interval_minutes")))
    }

    // generate ticker and channel for messages
    ticker := time.NewTicker(time.Duration(interval) * time.Minute)
    quitChan := make(chan bool)

    var wg sync.WaitGroup
    // add to waitgroup to prevent go routine from closing
    wg.Add(1)

    go func() {
        for {
            select {
            case <- ticker.C:
                log.Info("starting new data clearing job...")
                start := time.Now()

                if err := clearData(retentionPeriod); err != nil {
                    log.Error(fmt.Errorf("unable to clear data: %+v", err))
                }
                // log total time elapsed to process job
                elapsed := time.Now().Sub(start)
                log.Info(fmt.Sprintf("finished clearing job. took %fs to process", elapsed.Seconds()))
            case <- quitChan:
                // stop ticker and add to waitgroup
                ticker.Stop()
                wg.Done()
                return
            }
        }
    }()

    wg.Wait()
    log.Info("stopping data clearer...")
}