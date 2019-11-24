package alertmanager

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/utils"
	"sync"
	"time"
)

type AlertManager struct {
	sync.RWMutex
	Enable            bool
	ticker            time.Ticker
	reminder          time.Ticker
	vigieURL          string
	vigieInstanceName string
	email             email
	alrtList          alrtList
	hooks             map[string]hook
}

type alrtList struct {
	Testsuites map[int64]*teststruct.TestSuite
	anyChanges bool
}

type alertType bool

const (
	normal   alertType = false
	reminder alertType = true
)

type hook interface {
	send(tamsg teststruct.TotalAlertMessage, at alertType) error
	name() string
}

var AM AlertManager

func InitAlertManager(vConfAlerting ConfAlerting, vigieInstName, vigieURL string) error {

	AM.Lock()
	AM.Enable = vConfAlerting.Enable
	AM.vigieInstanceName = vigieInstName
	AM.vigieURL = vigieURL

	if vConfAlerting.Enable == true {

		errHook := AM.loadHooks(vConfAlerting)
		if errHook != nil {
			return errHook
		}

		AM.alrtList.Testsuites = make(map[int64]*teststruct.TestSuite, 0)

		if vConfAlerting.Interval == 0 {
			AM.ticker = *time.NewTicker(time.Second * 5)
		} else {

			if vConfAlerting.Interval <= time.Millisecond {
				return fmt.Errorf("AlertManager cannot check so frequently: interval cannot be < 1ms")
			}

			AM.ticker = *time.NewTicker(vConfAlerting.Interval)

		}

		if vConfAlerting.Reminder == 0 {
			AM.reminder = *time.NewTicker(time.Hour * 4)
		} else {

			if vConfAlerting.Reminder <= time.Second {
				return fmt.Errorf("AlertManager reminder cannot be set so frequently: interval cannot be =< 1s")
			}

			AM.reminder = *time.NewTicker(vConfAlerting.Reminder)
		}

		utils.Log.WithFields(logrus.Fields{
			"component": "alerting",
			"status":    "enable",
		}).Infof(fmt.Sprintf("AlertManager is set to: %t", AM.Enable))
	} else {
		utils.Log.WithFields(logrus.Fields{
			"component": "alerting",
			"status":    "disable",
		}).Infof(fmt.Sprintf("AlertManager is set to: %t", AM.Enable))
	}
	AM.Unlock()

	go AM.run()

	return nil
}

func (am *AlertManager) loadHooks(vigieConf ConfAlerting) error {

	// Hooks ---
	AM.hooks = make(map[string]hook, 0)

	// Add Discord if present
	if vigieConf.Discord.Hook != "" {

		da := discordAlert{webhookURL: vigieConf.Discord.Hook}
		am.hooks["discord"] = &da

	}

	// Add Email if present
	if vigieConf.Email.To != "" {
		ea := email{
			To:       vigieConf.Email.To,
			From:     vigieConf.Email.From,
			Username: vigieConf.Email.Username,
			Password: vigieConf.Email.Password,
			SMTP:     vigieConf.Email.SMTP,
			Port:     vigieConf.Email.Port,
		}
		am.hooks["email"] = &ea
	}

	if len(AM.hooks) == 0 {
		utils.Log.WithFields(logrus.Fields{
			"component": "alerting",
			"status":    "enable",
		}).Warnf("Alerting is enable but no hooks have been loaded.", AM.Enable)
	}
	return nil
}

// __________________________________________________________________________________________
// AddToAlertList add the task into the AlertManager
// If a TxStep has no leafs => delete this TxStep on AM
func (am *AlertManager) AddToAlertList(task teststruct.Task) error {

	ts := task.TestSuite
	tc := task.TestCase
	tstep := task.TestStep

	is_success := tstep.GetStatus() == teststruct.Success
	task.RLockAll()
	am.Lock()
	am.alrtList.anyChanges = true

	if is_success {
		// Is this OK Tstep present ?
		if _, ok := am.alrtList.Testsuites[ts.ID]; ok {
			if _, ok := am.alrtList.Testsuites[ts.ID].TestCases[tc.ID]; ok {
				if _, ok := am.alrtList.Testsuites[ts.ID].TestCases[tc.ID].TestSteps[tstep.ID]; ok {
					delete(am.alrtList.Testsuites[ts.ID].TestCases[tc.ID].TestSteps, tstep.ID)
				}
				// Is this OK Tstep the last present ? => Then Delete Parent
				if len(am.alrtList.Testsuites[ts.ID].TestCases[tc.ID].TestSteps) == 0 {
					delete(am.alrtList.Testsuites[ts.ID].TestCases, tc.ID)
				}
			}
			if len(am.alrtList.Testsuites[ts.ID].TestCases) == 0 {
				delete(am.alrtList.Testsuites, ts.ID)
			}
		}
	} else {

		// If Testsuite is not register, then no TC nor TStep is.
		// Add them all
		if _, here := am.alrtList.Testsuites[ts.ID]; !here {

			am.alrtList.Testsuites[ts.ID] = ts

			am.alrtList.Testsuites[ts.ID].TestCases = make(map[int64]*teststruct.TestCase, 0)
			am.alrtList.Testsuites[ts.ID].TestCases[tc.ID] = tc
			am.alrtList.Testsuites[ts.ID].TestCases[tc.ID].TestSteps = make(map[int64]*teststruct.TestStep, 0)
			am.alrtList.Testsuites[ts.ID].TestCases[tc.ID].TestSteps[tstep.ID] = tstep

		} else {
			// This TestSuites is already register => Add TC (if not register)
			if _, here := am.alrtList.Testsuites[ts.ID].TestCases[tc.ID]; !here {

				am.alrtList.Testsuites[ts.ID].TestCases[tc.ID] = tc
				am.alrtList.Testsuites[ts.ID].TestCases[tc.ID].TestSteps[tstep.ID] = tstep

			} else {
				// This TC is already register => Add TStep (if not register)
				am.alrtList.Testsuites[ts.ID].TestCases[tc.ID].TestSteps[tstep.ID] = tstep

			}
		}
	}
	am.Unlock()
	task.RUnlockAll()
	return nil
}

// sendHooks send the AlertMessage to every notifications services registered
func (am *AlertManager) sendHooks(amsg *teststruct.TotalAlertMessage, at alertType) {

	for _, hook := range am.hooks {

		err := hook.send(*amsg, at)
		if err != nil {
			utils.Log.WithFields(logrus.Fields{
				"component": "alerting",
				"target":    hook.name(),
			}).Errorf("Sending the %s alert failed : %s", hook.name(), err.Error())
		}
	}

	if at != reminder {
		am.resetChangeState()
	}
}

func (am *AlertManager) IsEnable() (enable bool) {

	am.RLock()
	enable = am.Enable
	am.RUnlock()
	return enable

}

func (am *AlertManager) anyChange() (change bool) {

	am.RLock()
	change = am.alrtList.anyChanges
	am.RUnlock()
	return change

}

func (am *AlertManager) resetChangeState() {

	am.RLock()
	am.alrtList.anyChanges = false
	am.RUnlock()

}
