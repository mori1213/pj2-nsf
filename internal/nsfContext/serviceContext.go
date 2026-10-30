package nsfContext

import (
	"encoding/json"
	"fmt"
	"nsf/internal/db"
	"nsf/logger"
)

// TODO: You may need to add some fields
type serviceContextIE struct {
	db.DbIf

	*logger.NsfLogger
}

// TODO: You may need to add some fields
type serviceContext struct {
	db.DbIf

	*logger.NsfLogger
}

// serviceRecord stores the NF information associated with a service.
type serviceRecord struct {
	NfName     string `json:"nf_name"`
	SbiAddress string `json:"sbi_address"`
	SbiPort    int    `json:"sbi_port"`
}

// TODO: Finish the initialization of the serviceContext instance
func newServiceContext(serviceContextIE *serviceContextIE) *serviceContext {

	return &serviceContext{
		DbIf: serviceContextIE.DbIf,

		NsfLogger: serviceContextIE.NsfLogger,
	}
}

// TODO: Finish the release of the serviceContext instance
func (s *serviceContext) release() {
	s.NsfLogger.SvcLog.Infoln("Release serviceContext...")

	s.NsfLogger.SvcLog.Infoln("serviceContext released")
}

// TODO: Finish the implementation of the registerServices method with an error return type
func (s *serviceContext) registerServices(nfName, sbiAddr string, sbiPort int, services []string) error {
	s.NsfLogger.SvcLog.Tracef("Registering %s services", nfName)

	record := serviceRecord{
		NfName:     nfName,
		SbiAddress: sbiAddr,
		SbiPort:    sbiPort,
	}

	value, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode services for NF %s: %w", nfName, err)
	}

	data := make(map[string]string, len(services))
	for _, service := range services {
		data[service] = string(value)
	}

	if err := s.DbIf.SaveAll(data); err != nil {
		return fmt.Errorf("save services for NF %s: %w", nfName, err)
	}

	return nil

}

// TODO: Finish the implementation of the deregisterServices method with an error return type
func (s *serviceContext) deregisterServices(nfName string, services []string) error {
	s.NsfLogger.SvcLog.Tracef("De-registering %s services", nfName)

	for _, service := range services {
		owner, _, _, err := s.discoverService(service)
		if err != nil {
			return err
		}
		if owner != nfName {
			return fmt.Errorf("service %s does not belong to NF %s", service, nfName)
		}
	}

	for _, service := range services {
		if err := s.DbIf.Remove(service); err != nil {
			return fmt.Errorf("remove service %s: %w", service, err)
		}
	}

	return nil
}

// TODO: Finish the implementation of the discoverService method with these return values:
// (string, string, int, error) representing the NF name, SBI address, SBI port, and an error if any
func (s *serviceContext) discoverService(service string) (string, string, int, error) {
	s.NsfLogger.SvcLog.Tracef("Discovering service %s", service)
	exists, err := s.DbIf.Exist(service)
	if err != nil {
		return "", "", 0, fmt.Errorf("check service %s: %w", service, err)
	}
	if !exists {
		return "", "", 0, fmt.Errorf("service %s not found", service)
	}

	value, err := s.DbIf.Load(service)
	if err != nil {
		return "", "", 0, fmt.Errorf("load service %s: %w", service, err)
	}

	var record serviceRecord

	if err := json.Unmarshal([]byte(value), &record); err != nil {
		return "", "", 0, fmt.Errorf("decode service %s: %w", service, err)
	}

	return record.NfName, record.SbiAddress, record.SbiPort, nil
}
