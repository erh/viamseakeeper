package viamseakeeper

import(
	"testing"

	"go.viam.com/test"
)

func TestParseMessage(t *testing.T) {

   s := "{\"active_alarm_history\":0,\"angle_calibration_complete\":false,\"battery_displayed\":0,\"battery_segments\":63,\"battery_voltage\":25.399999618530273,\"boat_roll_angle\":0.023101806640625,\"brightness_level\":0,\"checksums\":[{\"fname\":\"app_space\",\"hash\":\"fff1d698\"},{\"fname\":\"calib_space\",\"hash\":\"7d152eb8\"}],\"display\":\"7.43.0.36\",\"drive1\":\"20020006\",\"drive2\":\"4.00 / 4\",\"drive_current\":0,\"drive_temperature\":\"95.0° F\",\"enclosure_pressure\":\"65535\",\"flywheel_crt_selected_speed\":0,\"flywheel_max_speed\":5150,\"flywheel_min_speed\":4000,\"flywheel_speed\":0,\"gcm\":\"11.131\",\"graph_is_dial_left\":1,\"gyro_angle\":44.800003051757812,\"has_leds\":false,\"imu\":\"2.72\",\"language\":0,\"lic\":16777215,\"model\":\"40\",\"night_mode\":0,\"override_available\":true,\"override_list\":{\"brake_override\":false,\"glycol_override\":false,\"seawater_override\":false},\"power_available\":1,\"power_enabled\":0,\"progress_bar_available\":false,\"progress_bar_percentage\":0,\"pva_3g4\":\"4279\",\"pva_4g5\":\"8\",\"pva_5g6\":\"0\",\"pva_6g7\":\"0\",\"pva_7g\":\"0\",\"pva_g1\":\"5136\",\"pvmo\":530,\"run_hours\":1167,\"sea_hours\":958,\"seakeeper_instance\":1,\"serial\":\"234-0033\",\"stabilize_available\":false,\"stabilize_enabled\":0,\"synch_leds_available\":false,\"synch_leds_enabled\":false,\"tcml\":\"9213\",\"temp_is_celsius\":0}"

	sp, _, err := decodeMessage([]byte(s))
	test.That(t, err, test.ShouldBeNil)
	test.That(t, sp.BatteryVoltage, test.ShouldAlmostEqual, 25.399, .1)
}
	
