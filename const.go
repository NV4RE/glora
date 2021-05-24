package glora

type Mode byte
type Register byte
type PAConfig byte

const (
	RegFifo               Register = 0x00
	RegOpMode             Register = 0x01
	RegFrfMsb             Register = 0x06
	RegFrfMid             Register = 0x07
	RegFrfLsb             Register = 0x08
	RegPaConfig           Register = 0x09
	RegOcp                Register = 0x0b
	RegLna                Register = 0x0c
	RegFifoAddrPtr        Register = 0x0d
	RegFifoTxBaseAddr     Register = 0x0e
	RegFifoRxBaseAddr     Register = 0x0f
	RegFifoRxCurrentAddr  Register = 0x10
	RegIrqFlags           Register = 0x12
	RegRxNbBytes          Register = 0x13
	RegPktSnrValue        Register = 0x19
	RegPktRssiValue       Register = 0x1a
	RegRssiValue          Register = 0x1b
	RegModemConfig1       Register = 0x1d
	RegModemConfig2       Register = 0x1e
	RegPreambleMsb        Register = 0x20
	RegPreambleLsb        Register = 0x21
	RegPayloadLength      Register = 0x22
	RegModemConfig3       Register = 0x26
	RegFreqErrorMsb       Register = 0x28
	RegFreqErrorMid       Register = 0x29
	RegFreqErrorLsb       Register = 0x2a
	RegRssiWideBand       Register = 0x2c
	RegDetectionOptimize  Register = 0x31
	RegInvertIQ           Register = 0x33
	RegDetectionThreshold Register = 0x37
	RegSyncWord           Register = 0x39
	RegInvertIQ2          Register = 0x3b
	RegDioMapping1        Register = 0x40
	RegVersion            Register = 0x42
	RegPaDac              Register = 0x4d
)

const (
	ModeLongRange    Mode = 0x80
	ModeSleep        Mode = 0x00
	ModeStandby      Mode = 0x01
	ModeTx           Mode = 0x03
	ModeRxContinuous Mode = 0x05
	ModeRxSingle     Mode = 0x06
)

const (
	PABoost PAConfig = 0x80
)

const (
	IrqTxDoneMask          byte   = 0x08
	IrqPayloadCrcErrorMask byte   = 0x20
	IrqRxDoneMask          byte   = 0x40
	RfMidBandThreshold     uint64 = 525e6
	RssiOffsetHfPort       uint   = 157
	RssiOffsetLfPort       uint   = 164
	MaxPktLength           uint   = 255
)
