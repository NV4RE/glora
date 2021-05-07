package glora

import (
	"errors"
	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
	"time"
)

var (
	ErrGetVersion = errors.New("version not matched")
)

type Message struct {
	data []byte
	rssi int
	snr  float64
}

type Lora struct {
	SPI                spi.Conn
	DI0                gpio.PinIO
	implicitHeaderMode bool
	spreadingFactor    uint8
	signalBandwidth    uint64
	frequency          uint64
	codingRate         float64
	preambleLength     uint16
	syncWord           byte
	txPower            uint8
	crc                bool
	MessageCh          chan Message
}

func NewLora(spiDev, di0 string) (*Lora, error) {
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	_, err := driverreg.Init()
	if err != nil {
		return nil, err
	}

	p, err := spireg.Open(spiDev)
	if err != nil {
		return nil, err
	}

	c, err := p.Connect(physic.MegaHertz, spi.Mode3, 8)
	if err != nil {
		return nil, err
	}

	if _, err := driverreg.Init(); err != nil {
		return nil, err
	}

	dio0 := gpioreg.ByName(di0)
	if p == nil {
		return nil, errors.New("failed to find GPIO6")
	}

	if err := dio0.In(gpio.PullNoChange, gpio.RisingEdge); err != nil {
		return nil, err
	}

	return &Lora{
		SPI:                c,
		DI0:                dio0,
		implicitHeaderMode: false,
		spreadingFactor:    7,
		signalBandwidth:    125e3,
		frequency:          915e6,
		codingRate:         4 / 5,
		preambleLength:     8,
		syncWord:           0x12,
		txPower:            17,
		crc:                true,
	}, nil
}

func (l *Lora) Config() error {
	v, err := l.GetVersion()
	if err != nil {
		return err
	}
	if v != 0x12 {
		return ErrGetVersion
	}

	err = l.SetMode(ModeSleep)
	if err != nil {
		return err
	}

	err = l.SetFrequency(l.frequency)
	if err != nil {
		return err
	}

	//err = l.SetSpreadingFactor(l.spreadingFactor)
	//if err != nil {
	//	return err
	//}
	//
	//err = l.SetSignalBandwidth(l.signalBandwidth)
	//if err != nil {
	//	return err
	//}
	//
	//err = l.SetCodingRate(l.codingRate)
	//if err != nil {
	//	return err
	//}
	//
	//err = l.SetCrc(l.crc)
	//if err != nil {
	//	return err
	//}

	err = l.WriteRegister(RegFifoTxBaseAddr, 0)
	if err != nil {
		return err
	}
	err = l.WriteRegister(RegFifoRxBaseAddr, 0)
	if err != nil {
		return err
	}

	err = l.SetLnaBoost(true)
	if err != nil {
		return err
	}

	err = l.WriteRegister(RegModemConfig3, 0x04)
	if err != nil {
		return err
	}

	err = l.SetTxPower(l.txPower)
	if err != nil {
		return err
	}

	return l.SetMode(ModeStandby)
}

func (l *Lora) GetVersion() (byte, error) {
	return l.ReadRegister(RegVersion)
}

func (l *Lora) SetMode(m Mode) error {
	return l.WriteRegister(RegOpMode, byte(ModeLongRange|m))
}

func (l *Lora) SetFrequency(frequency uint64) error {
	frf := (frequency << 19) / 32000000

	err := l.WriteRegister(RegFrfMsb, byte(frf>>16))
	if err != nil {
		return err
	}
	err = l.WriteRegister(RegFrfMid, byte(frf>>8))
	if err != nil {
		return err
	}
	return l.WriteRegister(RegFrfLsb, byte(frf>>0))
}

func (l *Lora) SetLnaBoost(boost bool) error {
	lna, err := l.ReadRegister(RegLna)
	if err != nil {
		return err
	}

	if boost {
		return l.WriteRegister(RegLna, lna|0x03)
	}
	return l.WriteRegister(RegLna, lna&0xfc)
}

func (l *Lora) SetSpreadingFactor(sf uint8) error {
	if sf < 6 {
		sf = 6
	} else if sf > 12 {
		sf = 12
	}

	var detectionOptimize byte = 0xc3
	var detectionThreshold byte = 0x0a
	if sf == 6 {
		detectionOptimize = 0xc5
		detectionThreshold = 0x0c
	}

	err := l.WriteRegister(RegDetectionOptimize, detectionOptimize)
	if err != nil {
		return err
	}

	err = l.WriteRegister(RegDetectionThreshold, detectionThreshold)
	if err != nil {
		return err
	}

	mc, err := l.ReadRegister(RegModemConfig2)
	if err != nil {
		return err
	}

	return l.WriteRegister(RegModemConfig2, (mc&0x0f)|(sf<<4))
}

func (l *Lora) SetTxPower(power uint8) error {
	if power < 2 {
		power = 2
	} else if power > 17 {
		power = 17
	}
	return l.WriteRegister(RegPaConfig, power)
}

func (l *Lora) SetSignalBandwidth(bandwidth uint64) error {
	var bw byte
	if bandwidth <= 7.8e3 {
		bw = 0
	} else if bandwidth <= 10.4e3 {
		bw = 1
	} else if bandwidth <= 15.6e3 {
		bw = 2
	} else if bandwidth <= 20.8e3 {
		bw = 3
	} else if bandwidth <= 31.25e3 {
		bw = 4
	} else if bandwidth <= 41.7e3 {
		bw = 5
	} else if bandwidth <= 62.5e3 {
		bw = 6
	} else if bandwidth <= 125e3 {
		bw = 7
	} else if bandwidth <= 250e3 {
		bw = 8
	} else { // bandwidth <= 250E3
		bw = 9
	}

	mc, err := l.ReadRegister(RegModemConfig1)
	if err != nil {
		return err
	}

	return l.WriteRegister(RegModemConfig1, (mc&0x0f)|(bw<<4))
}

func (l *Lora) SetPreambleLength(length uint16) error {
	err := l.WriteRegister(RegPreambleMsb, byte(length>>8))
	if err != nil {
		return err
	}
	err = l.WriteRegister(RegPreambleLsb, byte(length>>0))
	if err != nil {
		return err
	}
	l.preambleLength = length
	return nil
}

func (l *Lora) SetCodingRate(cr float64) error {
	var denominator byte
	if cr <= (4 / 5) {
		denominator = 5
	} else if cr <= (4 / 6) {
		denominator = 6
	} else if cr <= (4 / 7) {
		denominator = 7
	} else { // cr <= 4/8
		denominator = 8
	}

	mc, err := l.ReadRegister(RegModemConfig1)
	if err != nil {
		return err
	}
	codingRate := denominator - 4

	l.codingRate = cr
	return l.WriteRegister(RegModemConfig1, (mc&0xf1)|(codingRate<<1))
}

func (l *Lora) SetCrc(crc bool) error {
	mc, err := l.ReadRegister(RegModemConfig2)
	if err != nil {
		return err
	}
	l.crc = crc
	if crc {
		return l.WriteRegister(RegModemConfig2, mc|0x04)
	}
	return l.WriteRegister(RegModemConfig2, mc&0xfb)
}

func (l *Lora) GetFIFO() ([]byte, error) {
	rxAddr, err := l.ReadRegister(RegFifoRxCurrentAddr)
	if err != nil {
		return nil, err
	}

	err = l.WriteRegister(RegFifoAddrPtr, rxAddr)
	if err != nil {
		return nil, err
	}

	var nbBytes byte
	if l.implicitHeaderMode {
		nbBytes, err = l.ReadRegister(RegPayloadLength)
		if err != nil {
			return nil, err
		}
	} else {
		nbBytes, err = l.ReadRegister(RegRxNbBytes)
		if err != nil {
			return nil, err
		}
	}

	return l.ReadRegisterBytes(RegFifo, int(nbBytes))
}

func (l *Lora) GetRssi() (uint8, error) {
	rssi, err := l.ReadRegister(RegPktRssiValue)
	if err != nil {
		return 0, err
	}

	if l.frequency < 868e6 {
		return rssi - 164, nil
	}
	return rssi - 157, nil
}

func (l *Lora) GetSNR() (float64, error) {
	snr, err := l.ReadRegister(RegPktSnrValue)
	if err != nil {
		return 0, err
	}
	return float64(snr) * 0.25, nil
}

func (l *Lora) ClearIrqFlags() (byte, error) {
	irq, err := l.ReadRegister(RegIrqFlags)
	if err != nil {
		return 0, err
	}

	return irq, l.WriteRegister(RegIrqFlags, irq)
}

func (l *Lora) Receive(timeout time.Duration) ([]byte, error) {
	err := l.ExplicitHeaderMode()
	if err != nil {
		return nil, err
	}

	_, err = l.ClearIrqFlags()
	if err != nil {
		return nil, err
	}

	err = l.WriteRegister(RegRxNbBytes, 0)
	if err != nil {
		return nil, err
	}

	err = l.WriteRegister(RegDioMapping1, 0x00)
	if err != nil {
		return nil, err
	}

	err = l.SetMode(ModeRxContinuous)
	if err != nil {
		return nil, err
	}
	defer l.SetMode(ModeStandby)

	raise := l.DI0.WaitForEdge(timeout)
	if raise {
		return l.GetFIFO()
	}
	return nil, nil
}

func (l *Lora) Transmit(b []byte, t time.Duration) error {
	err := l.WriteRegister(RegFifoAddrPtr, 0)
	if err != nil {
		return err
	}

	err = l.WriteRegister(RegPayloadLength, uint8(len(b)))
	if err != nil {
		return err
	}

	err = l.WriteRegister(RegFifo, b...)
	if err != nil {
		return err
	}

	return nil
}

func (l *Lora) ExplicitHeaderMode() error {
	mc1, err := l.ReadRegister(RegModemConfig1)
	if err != nil {
		return err
	}
	l.implicitHeaderMode = false
	return l.WriteRegister(RegModemConfig1, mc1&0xfe)
}

func (l *Lora) ImplicitHeaderMode() error {
	mc1, err := l.ReadRegister(RegModemConfig1)
	if err != nil {
		return err
	}
	l.implicitHeaderMode = true
	return l.WriteRegister(RegModemConfig1, mc1&0x01)
}

func (l *Lora) ReadRegister(a Register) (byte, error) {
	b := []byte{byte(a) & 0x7f, 0x00}
	read := make([]byte, len(b))
	err := l.SPI.Tx(b, read)
	if err != nil {
		return 0, err
	}
	return read[1], nil
}

func (l *Lora) ReadRegisterBytes(a Register, n int) ([]byte, error) {
	b := append([]byte{byte(a) & 0x7f, 0x00}, make([]byte, n)...)
	read := make([]byte, len(b))
	err := l.SPI.Tx(b, read)
	if err != nil {
		return read, err
	}
	return read[1:], nil
}

func (l *Lora) WriteRegister(a Register, v ...byte) error {
	return l.SPI.Tx(append([]byte{byte(a) | 0x80}, v...), make([]byte, 2))
}
