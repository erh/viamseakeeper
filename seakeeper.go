package viamseakeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse/paho.mqtt.golang"

	"go.viam.com/rdk/data"
	"go.viam.com/rdk/logging"
)

type Status struct {
	BatteryVoltage float64 `json:"battery_voltage"`
	BoatRollAngle float64 `json:"boat_roll_angle"`
	SeaHours float64 `json:"sea_hours"`
	
	DriveCurrent float64 `json:"drive_current"`

	DriveTemperature string `json:"drive_temperature"`

	StabstabilizeEnabled float64 `json:"stabilize_enabled"`
	StabilizeAvailable bool `json:"stabilize_available"`
	PowerAvailable float64 `json:"power_available"`
	PowerEnabled float64 `json:"power_enabled"`
}

func NewSeakeeper(host string, logger logging.Logger) (*Seakeeper, error) {
	s := &Seakeeper{host: host, logger: logger}

	return s, nil
}

type Seakeeper struct {
	host string
	logger logging.Logger
	
	client mqtt.Client

	lastStatus map[string]interface{}
	lastStatusParsed Status
	lastStatusTime time.Time
}


// Not thread safe
func (s *Seakeeper) Start() error {
	if s.client != nil {
		return nil
	}

	//mqtt.DEBUG = s.logger
	//mqtt.ERROR = s.logger
	
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("ws://%s:9001", s.host))
	opts.SetAutoReconnect(true)

	s.client = mqtt.NewClient(opts)
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	f := func(client mqtt.Client, msg mqtt.Message) {
		sp, m, err := decodeMessage(msg.Payload())
		if err != nil {
			s.logger.Errorf("error decoding message %v %v", msg, err)
			return
		}
		s.lastStatus = m
		s.lastStatusParsed = sp
		s.lastStatusTime = time.Now()
	}
	
	if token := s.client.Subscribe("seakeeper/status/1", 0, f); token.Wait() && token.Error() != nil {
		s.client.Disconnect(0)
		s.client = nil
		return token.Error()
	}

	return nil
}

func (s *Seakeeper) Power(on bool) error {
	err := tooOld(nil, s.lastStatusTime)
	if err != nil {
		return err
	}
	if s.lastStatusParsed.PowerEnabled > 0 {
		return nil
	}
	if on && s.lastStatusParsed.PowerAvailable < 1 {
		return fmt.Errorf("trying to turn power on and not available")

	}
	panic(1)
}

func (s *Seakeeper) Enable(on bool) error {
	err := tooOld(nil, s.lastStatusTime)
	if err != nil {
		return err
	}
	if s.lastStatusParsed.StabstabilizeEnabled > 0 {
		return nil
	}
	if on && !s.lastStatusParsed.StabilizeAvailable {
		return fmt.Errorf("trying to enable on and not available")

	}
	panic(1)
}

func (s *Seakeeper) LastStatus() Status {
	return s.lastStatusParsed
}

func (s *Seakeeper) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	err := tooOld(extra, s.lastStatusTime)
	if err != nil {
		return nil, err
	}
	return s.lastStatus, nil
}

func (s *Seakeeper) Close(ctx context.Context) error {
	if s.client != nil {
		s.client.Disconnect(250)
		s.client = nil
	}
	return nil
}

func decodeMessage(b []byte) (Status, map[string]interface{}, error) {
	s := Status{}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return s, nil, err
	}
	m := map[string]interface{}{}
	err = json.Unmarshal(b, &m)
	return s, m, err
}

// taken from viamboat
func isFromDataCapture(extra map[string]interface{}) bool {
	if extra == nil {
		return false
	}

	return extra[data.FromDMString] == true
}

func tooOld(extra map[string]interface{}, lastUpdate time.Time) error {
	if time.Since(lastUpdate) < time.Minute {
		return nil
	}

	if isFromDataCapture(extra) {
		// we're from data capture
		// since data is too old, just don't store anything or log
		return data.ErrNoCaptureToStore
	}

	return fmt.Errorf("lastUpdate update too old: %v (%v)", lastUpdate, extra)
}
