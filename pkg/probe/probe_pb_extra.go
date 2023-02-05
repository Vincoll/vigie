package probe

type ProbeNotValidated interface {
	ValidateAndInit() error
}

type ProbeNotVal any

func (x ProbeNotVal) name() {

}
