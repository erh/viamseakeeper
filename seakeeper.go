package viamseakeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eclipse/paho.mqtt.golang"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/data"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var family = resource.ModelNamespace("erh").WithFamily("viamseakeeper")

var Model = family.WithModel("seakeeper")

func init() {
	resource.RegisterComponent(
		sensor.API,
		Model,
		resource.Registration[sensor.Sensor, resource.NoNativeConfig]{
			Constructor: newSeakeeperSensor,
		})
}

func newSeakeeperSensor(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	host := config.Attributes.String("host")
	if host == "" {
		return nil, fmt.Errorf("need to specify host")
	}

	s, err := NewSeakeeper(host, logger)
	if err != nil {
		return nil, err
	}

	s.name = config.ResourceName()

	err = s.Start()
	if err != nil {
		return nil, err
	}

	return s, nil
}

type Status struct {
	BatteryVoltage float64 `json:"battery_voltage"`
	BoatRollAngle  float64 `json:"boat_roll_angle"`
	SeaHours       float64 `json:"sea_hours"`

	DriveCurrent float64 `json:"drive_current"`

	DriveTemperature string `json:"drive_temperature"`

	ProgressBar float64 `json:"progress_bar_percentage"`

	StabilizeEnabled   float64 `json:"stabilize_enabled"`
	StabilizeAvailable bool    `json:"stabilize_available"`
	PowerAvailable     float64 `json:"power_available"`
	PowerEnabled       float64 `json:"power_enabled"`
}

func NewSeakeeper(host string, logger logging.Logger) (*Seakeeper, error) {
	s := &Seakeeper{host: host, logger: logger}

	return s, nil
}

type Seakeeper struct {
	resource.AlwaysRebuild

	name resource.Name

	host   string
	logger logging.Logger

	client mqtt.Client

	lastStatus       map[string]interface{}
	lastStatusParsed Status
	lastStatusTime   time.Time
}

type myLogAdapter struct {
	logger logging.Logger
}

func (l *myLogAdapter) Println(v ...interface{}) {
	l.logger.Error(v...)
}

func (l *myLogAdapter) Printf(format string, v ...interface{}) {
	l.logger.Errorf(format, v...)
}

// Not thread safe
func (s *Seakeeper) Start() error {
	if s.client != nil {
		return nil
	}

	//mqtt.DEBUG = s.logger
	mqtt.ERROR = &myLogAdapter{s.logger}

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
	if on && s.lastStatusParsed.PowerEnabled > 0 {
		return nil
	}
	if on && s.lastStatusParsed.PowerAvailable < 1 {
		return fmt.Errorf("trying to turn power on and not available")

	}

	m := map[string]interface{}{}
	if on {
		m["power"] = 1.0
	} else {
		m["power"] = 0.0
	}
	return s.sendRequest(m)
}

// var textR = '{"stabilize":1}';
func (s *Seakeeper) Enable(on bool) error {
	err := tooOld(nil, s.lastStatusTime)
	if err != nil {
		return err
	}
	if on && s.lastStatusParsed.StabilizeEnabled > 0 {
		return nil
	}
	if on && !s.lastStatusParsed.StabilizeAvailable {
		return fmt.Errorf("trying to enable on and not available")

	}

	m := map[string]interface{}{}
	if on {
		m["stabilize"] = 1.0
	} else {
		m["stabilize"] = 0.0
	}
	return s.sendRequest(m)

}

func (s *Seakeeper) sendRequest(m map[string]interface{}) error {
	if s.client == nil {
		return fmt.Errorf("no client")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	fmt.Printf("yo %v\n", string(b))
	token := s.client.Publish("seakeeper/request/1", 0, false, string(b))
	if !token.WaitTimeout(time.Second * 30) {
		return fmt.Errorf("timed out sending request %v", m)
	}
	return token.Error()
}

func (s *Seakeeper) LastStatus() Status {
	return s.lastStatusParsed
}

func (s *Seakeeper) LastStatusTime() time.Time {
	return s.lastStatusTime
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

func (s *Seakeeper) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (s *Seakeeper) Name() resource.Name {
	return s.name
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
