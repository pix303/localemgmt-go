package aggregate

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// LocaleItemAggregateLocalPersisterOnFile is the implementation of the LocaleItemAggregatePersister interface for persist the aggregate on local file system
type LocaleItemAggregateLocalPersisterOnFile struct{}

const LocaleItemAggregateLocalFolder = "./localeitems-bucket"

func (persister *LocaleItemAggregateLocalPersisterOnFile) Persist(aggregate LocaleItemAggregate) error {
	slog.Info("start persisting aggregate", slog.String("aggregateId", aggregate.AggregateID))
	aggregateDir, err := os.Open(LocaleItemAggregateLocalFolder)
	if err != nil {
		err = os.Mkdir(LocaleItemAggregateLocalFolder, 0755)
		if err != nil {
			return fmt.Errorf("error on create local folder for aggregates: %s", err.Error())
		}
		aggregateDir, _ = os.Open(LocaleItemAggregateLocalFolder)
	}

	resjson, err := json.Marshal(aggregate)
	if err != nil {
		slog.Warn("fail to marshal aggregate", slog.String("error", err.Error()))
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s.json", aggregateDir.Name(), aggregate.AggregateID), resjson, 0755)
	if err != nil {
		slog.Warn("fail to write aggregate file", slog.String("error", err.Error()))
		return err
	}
	return nil
}
