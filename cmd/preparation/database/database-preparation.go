package database

type DatabasePreparation interface {
	Preapare() func
}
