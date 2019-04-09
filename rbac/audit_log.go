package rbac

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"

	"github.com/searchlight/auditsink-to-rbac/system"
)

var engine *xorm.Engine

// Start the Postgres Database
func StartXormEngine() {
	var err error
	connStr := "user=masud password=masud123 host=127.0.0.1 port=5432 dbname=auditsink sslmode=disable"

	if engine, err = xorm.NewEngine("postgres", connStr); err != nil {
		log.Fatalln(err)
	}
	logFile, err := os.Create("auditsink-database.log")
	if err != nil {
		log.Fatalln(err)
	}
	logger := xorm.NewSimpleLogger(logFile)
	logger.ShowSQL(true)
	engine.SetLogger(logger)

	if engine.TZLocation, err = time.LoadLocation("Asia/Dhaka"); err != nil {
		log.Fatalln(err)
	}
}

func AddNewResource(event system.Event) error {
	auditLog := &system.AuditLog{
		EventID:      event.AuditID,
		ClusterUUID:  event.ClusterUUID,
		ResourceUUID: event.ResourceUUID,
		ResourceName: event.ResourceName,

		ResourceGroup:   event.ResourceGroup,
		ResourceVersion: event.ResourceVersion,
		ResourceKind:    event.ResourceKind,

		CreateTimestamp: event.StageTimestamp.Time,
		CreatedBy:       event.Username,
	}
	session := engine.NewSession()
	defer session.Close()
	if _, err := session.Insert(auditLog); err != nil {
		if err = session.Rollback(); err != nil {
			return err
		}
	}
	if err := session.Commit(); err != nil {
		return err
	}

	return nil
}

func UpdateExistingResource(event system.Event) error {
	auditLog := new(system.AuditLog)
	auditLog.ResourceUUID = event.ResourceUUID

	if exist, err := engine.Get(auditLog); !exist {
		return err
	}
	auditLog.DeleteTimestamp = event.StageTimestamp.Time
	auditLog.DeletedBy = event.Username

	session := engine.NewSession()
	defer session.Close()

	if _, err := session.ID(auditLog.ResourceUUID).Update(auditLog); err != nil {
		if err = session.Rollback(); err != nil {
			return err
		}
	}
	if err := session.Commit(); err != nil {
		return err
	}
	return nil
}

func SaveAuditLogToDatabase(event system.Event) error {
	StartXormEngine()
	defer engine.Close()

	if exist, _ := engine.IsTableExist(new(system.AuditLog)); !exist {
		if err := engine.CreateTables(new(system.AuditLog)); err != nil {
			return err
		}
	}
	if event.Verb == system.VerbCreate {
		return AddNewResource(event)
	} else if event.Verb == system.VerbDelete {
		return UpdateExistingResource(event)
	}
	return nil
}

func SaveAuditLogToDatabaseFromBytes(eventBytes []byte) error {
	eventList := system.EventList{}
	if err := json.Unmarshal(eventBytes, &eventList); err != nil {
		return err
	}
	event := eventList.Items[0]
	if event.ResponseCode != http.StatusCreated && event.ResponseCode != http.StatusOK {
		return nil
	}
	return SaveAuditLogToDatabase(event)
}
