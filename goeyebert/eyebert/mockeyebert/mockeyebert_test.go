package mockeyebert

import (
	"errors"
	"testing"
	"time"

	"github.com/allefr/eyebert/goeyebert/eyebert"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestClosePass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	eyebertMock.EXPECT().Close().Return(nil)

	err := eyebertMock.Close()
	assert.NoError(t, err)
}

func TestCloseFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	eyebertMock.EXPECT().Close().Return(errors.New("error while closing"))

	err := eyebertMock.Close()
	assert.Error(t, err)
	assert.Equal(t, "error while closing", err.Error())
}

func TestStartTestPass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	eyebertMock.EXPECT().StartTest().Return(nil)

	err := eyebertMock.StartTest()
	assert.NoError(t, err)
}

func TestStartTestFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	eyebertMock.EXPECT().StartTest().Return(errors.New("error while starting test"))

	err := eyebertMock.StartTest()
	assert.Error(t, err)
	assert.Equal(t, "error while starting test", err.Error())
}

func TestStopTestPass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	eyebertMock.EXPECT().StopTest().Return(nil)

	err := eyebertMock.StopTest()
	assert.NoError(t, err)
}

func TestStopTestFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	eyebertMock.EXPECT().StopTest().Return(errors.New("error while stopping test"))

	err := eyebertMock.StopTest()
	assert.Error(t, err)
	assert.Equal(t, "error while stopping test", err.Error())
}

func TestResetStatsPass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	eyebertMock.EXPECT().ResetStats().Return(nil)

	err := eyebertMock.ResetStats()
	assert.NoError(t, err)
}

func TestResetStatsFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	eyebertMock.EXPECT().ResetStats().Return(errors.New("error while resetting stats"))

	err := eyebertMock.ResetStats()
	assert.Error(t, err)
	assert.Equal(t, "error while resetting stats", err.Error())
}

func TestSetWaveLengthPass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expVal float32 = 1550.
	eyebertMock.EXPECT().SetWaveLength(expVal).Return(nil)

	err := eyebertMock.SetWaveLength(expVal)
	assert.NoError(t, err)
}

func TestSetWaveLengthFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expVal float32 = 1550.
	eyebertMock.EXPECT().SetWaveLength(expVal).Return(errors.New("error while setting wavelength"))

	err := eyebertMock.SetWaveLength(expVal)
	assert.Error(t, err)
	assert.Equal(t, "error while setting wavelength", err.Error())
}

func TestSetDataRatePass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expRateGbps float32 = 10.3125
	eyebertMock.EXPECT().SetDataRate(expRateGbps).Return(nil)

	err := eyebertMock.SetDataRate(expRateGbps)
	assert.NoError(t, err)
}

func TestSetDataRateFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expRateGbps float32 = 10.3125
	eyebertMock.EXPECT().SetDataRate(expRateGbps).Return(errors.New("error while setting datarate"))

	err := eyebertMock.SetDataRate(expRateGbps)
	assert.Error(t, err)
	assert.Equal(t, "error while setting datarate", err.Error())
}

func TestSetPatternPass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expPtn eyebert.Pattern = eyebert.PattPRBS31
	eyebertMock.EXPECT().SetPattern(expPtn).Return(nil)

	err := eyebertMock.SetPattern(expPtn)
	assert.NoError(t, err)
}

func TestSetPatternFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expPtn eyebert.Pattern = eyebert.PattPRBS31
	eyebertMock.EXPECT().SetPattern(expPtn).Return(errors.New("error while setting pattern"))

	err := eyebertMock.SetPattern(expPtn)
	assert.Error(t, err)
	assert.Equal(t, "error while setting pattern", err.Error())
}

func TestSetSFPtxEnablePass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expEn bool = true
	eyebertMock.EXPECT().SetSFPtxEnable(expEn).Return(nil)

	err := eyebertMock.SetSFPtxEnable(expEn)
	assert.NoError(t, err)
}

func TestSetSFPtxEnableFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expEn bool = true
	eyebertMock.EXPECT().SetSFPtxEnable(expEn).Return(errors.New("error while setting SetSFPtxEnable"))

	err := eyebertMock.SetSFPtxEnable(expEn)
	assert.Error(t, err)
	assert.Equal(t, "error while setting SetSFPtxEnable", err.Error())
}

func TestGetTesterInfoPass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	expInfo := eyebert.BERTester{
		Manufacturer: "Manuf",
		Model:        "eyebert",
		Version:      "v0.0",
	}
	eyebertMock.EXPECT().GetTesterInfo().Return(expInfo, nil)

	info, err := eyebertMock.GetTesterInfo()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, expInfo.Manufacturer, info.Manufacturer)
	assert.Equal(t, expInfo.Model, info.Model)
	assert.Equal(t, expInfo.Version, info.Version)
}

func TestGetTesterInfoFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expInfo eyebert.BERTester
	eyebertMock.EXPECT().GetTesterInfo().Return(expInfo, errors.New("error while getting tester info"))

	info, err := eyebertMock.GetTesterInfo()
	assert.Error(t, err)
	assert.Equal(t, "error while getting tester info", err.Error())
	assert.NotNil(t, info)
	assert.Equal(t, "", info.Manufacturer)
	assert.Equal(t, "", info.Model)
	assert.Equal(t, "", info.Version)
}

func TestGetSFPinfoPass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	expInfo := eyebert.SFPData{
		Vendor:       "sfp vendor",
		PartNum:      "123-4",
		WaveLengthNm: 1550.,
		DistanceKm:   120.,
		RxPow:        -24.3,
		TxPow:        1.3,
	}
	eyebertMock.EXPECT().GetSFPinfo().Return(expInfo, nil)

	info, err := eyebertMock.GetSFPinfo()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, expInfo.Vendor, info.Vendor)
	assert.Equal(t, expInfo.PartNum, info.PartNum)
	assert.Equal(t, expInfo.WaveLengthNm, info.WaveLengthNm)
	assert.Equal(t, expInfo.DistanceKm, info.DistanceKm)
	assert.Equal(t, expInfo.RxPow, info.RxPow)
	assert.Equal(t, expInfo.TxPow, info.TxPow)
}

func TestGetSFPinfoFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expInfo eyebert.SFPData
	eyebertMock.EXPECT().GetSFPinfo().Return(expInfo, errors.New("error while getting sfp info"))

	info, err := eyebertMock.GetSFPinfo()
	assert.Error(t, err)
	assert.Equal(t, "error while getting sfp info", err.Error())
	assert.NotNil(t, info)
	assert.Equal(t, "", info.Vendor)
	assert.Equal(t, "", info.PartNum)
	assert.Equal(t, float32(0.), info.WaveLengthNm)
	assert.Equal(t, float32(0.), info.DistanceKm)
	assert.Equal(t, float32(0.), info.RxPow)
	assert.Equal(t, float32(0.), info.TxPow)
}

func TestGetStatsPass(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	expInfo := eyebert.BERStats{
		Duration: 3 * time.Second,
		BER:      -1.,
		BitCnt:   123456789.,
		ErrCnt:   0.,
		RateGpbs: 10.3125,
		Pattern:  "PRBS31",
		Status:   "some issue",
	}
	eyebertMock.EXPECT().GetStats().Return(expInfo, nil)

	info, err := eyebertMock.GetStats()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, expInfo.Duration, info.Duration)
	assert.Equal(t, expInfo.BER, info.BER)
	assert.Equal(t, expInfo.BitCnt, info.BitCnt)
	assert.Equal(t, expInfo.ErrCnt, info.ErrCnt)
	assert.Equal(t, expInfo.RateGpbs, info.RateGpbs)
	assert.Equal(t, expInfo.Pattern, info.Pattern)
	assert.Equal(t, expInfo.Status, info.Status)
}

func TestGetStatsFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	eyebertMock := NewMockBERTDriver(mockCtrl)
	var expInfo eyebert.BERStats
	eyebertMock.EXPECT().GetStats().Return(expInfo, errors.New("error while getting stats"))

	info, err := eyebertMock.GetStats()
	assert.Error(t, err)
	assert.Equal(t, "error while getting stats", err.Error())
	assert.NotNil(t, info)
	assert.Equal(t, "", info.Status)
	assert.Equal(t, "", info.Pattern)
	assert.Equal(t, 0., info.BER)
	assert.Equal(t, 0., info.BitCnt)
	assert.Equal(t, 0., info.ErrCnt)
}
