import sys
import select
import math
import time
from datetime import datetime
import serial
import argparse
import json


pattLowPow = 'LOW_POW'
pattPRBS7  = 'PRBS7' # PRBS 2^7-1
pattPRBS9  = 'PRBS9' # PRBS 2^9-1
pattPRBS11 = 'PRBS11' # PRBS 2^11-1
pattPRBS15 = 'PRBS15' # PRBS 2^15-1
pattPRBS23 = 'PRBS23' # PRBS 2^23-1
pattPRBS31 = 'PRBS31' # PRBS 2^31-1
pattPRBS58 = 'PRBS58' # PRBS 2^58-1
pattPRBS63 = 'PRBS63' # PRBS 2^63-1
pattLoopBk = 'LOOPBACK' # data on the input is retransmitted on the output


class EyeBERT_MicroX(object):
  """EyeBERT_MicroX

  Attributes
  ----------
  serPort
    serial port path

  Methods
  -------
  getTesterInfo()
    returns dict of model and version of the tester
  getSFPinfo()
    returns dict with DDM info from the SFP
  getStats()
    returns dict with ongoing test stats
  setWaveLength(wl)
    sets wavelength in nm to use during the BER test
  setDataRate(dr)
    sets datarate in Gbps to use during the BER test
  setPattern(pattern)
    sets pattern to use during the BER test
  setSFPtxEnable(on)
    turns SFP tx On or Off
  resetStats()
    resets all test stats, effectively re-starting the BER test
  runQuickTest()
    runs quick test using datarates ranging from what the SFP supports
  Close()
    closes the serial port
  """

  _bertPatterns = {pattLowPow: "0",
                   pattPRBS7: "7",   
                   pattPRBS9: "9",
                   pattPRBS11: "1",
                   pattPRBS15: "5",
                   pattPRBS23: "2",
                   pattPRBS31: "3",
                   pattPRBS58: "8",
                   pattPRBS63: "6",
                   pattLoopBk: "L",
  }

  def __init__(self, port: str,
               datarate=None,
               pattern=None,
               ):
    """
    Parameters
    ----------
    port : str
      path to the serial port connected to the Tester
    datarate : optional
      datarate in Gbps. default taken from SFP
    pattern : optional
      BER pattern to use. default PRBS7
    """
    self.serPort = port
    self._ser = self._open()

    # BERT info
    self.getTesterInfo()

    if datarate is not None:
      self.setDataRate(datarate)
    if pattern is not None:
      self.setPattern(pattern)

    self.timeStart = datetime.now()
    self._prevBitCount = -1

    # reset stats
    self.resetStats()

  def _open(self):
    # note here baudrate can be anything
    ser = serial.Serial(port=self.serPort, baudrate=500000, timeout=1)
    if not ser.isOpen():
      print("can't open serial port! ('%s')" % self.serPort)
      exit(-1)

    # get rid of any garbage in the serial buffer
    ser.flushInput()
    ser.flushOutput()
    while ser.inWaiting():
      ser.read()

    return ser

  def Close(self):
    self._ser.close()

  def _write(self, cmd_string):
    while self._ser.inWaiting():
      self._ser.read()
    # do not use "=", it crashes the tester
    self._ser.write((cmd_string.replace("=", " ") + "\r\n").encode())
    time.sleep(.01)

  def _read(self, num=None):
    buff = bytearray([])
    while self._ser.inWaiting():
      buff.extend(self._ser.read())
    if num is not None and len(buff) < num:
      # try again after some time
      time.sleep(.5)
      while self._ser.inWaiting():
        buff.extend(self._ser.read())
    if len(buff) == 0:
      print("tester seems stuck! please power cycle it")
      self.Close()
      exit(-2)

    return buff

  def __print(self, s, dict_=None, overwrite=False):
    if self._noOutput:
      return
    if self._json:
      if dict_ is not None:
        s = json.dumps(dict_)
      else:
        return
    print(s, end='\r' if overwrite else '\n')

  def getTesterInfo(self):
    self._write("?")
    time.sleep(.02)
    b = (self._read()).decode()

    strn = b.split()
    if len(strn) >= 3:
      BERTmodel = strn[0] + " " + strn[1]
      BERTversion = strn[2].replace(":", "")
      status = 'ok'
    else:
      BERTmodel = None
      BERTversion = None
      status = "un-expected reply from tester"
    
    return {'model': BERTmodel, 'version': BERTversion, 'status': status}

  def getSFPinfo(self):
    self._write("sfp")
    time.sleep(.2)
    # total bytes to receive is around 2303, but only care about first part
    # still need to wait till all the data is received, or will fuck up other commands.. :(
    b = (self._read(2250)).decode()

    msg = None
    # this has multiple lines
    strn = b.split("\n")
    if len(strn) >= 12:
      for ln in strn:
        if "SFP Vendor" in ln:
          SFPvendor = ln.split()[2]
        if "Part Number" in ln:
          SFPpartNum = ln.split()[2]
        if "SN" in ln:
          SFPserial = ln.split()[1]
        if "Wavelength" in ln:
          SFPwaveLen = float(ln.split()[1])
        if "Temperature" in ln:
          SFPtemp = float(ln.split()[1])
        if "Rx Power" in ln:
          SFPrxPow = float(ln.split()[2])
        if "Tx Power" in ln:
          SFPtxPow = float(ln.split()[2])
          # no more useful lines after this, can exit
          break
        if "Media" in ln:
          SFPdistanceKm = float(ln.split()[2])
        if "Bit Rate" in ln:
          SFPminBitRateGbps = float(ln.split()[2])
          SFPmaxBitRateGbps = float(ln.split()[4])
    elif "No SFP Inserted" in strn:
      return {'status': "no SFP detected"}
    else:
      return {'status': "un-expected reply from tester"}

    return {'vendor': SFPvendor,
            'part-num': SFPpartNum,
            'serial': SFPserial,
            'wavelength': SFPwaveLen,
            'distance-km': SFPdistanceKm,
            'min-rate-gbps': SFPminBitRateGbps,
            'max-rate-gbps': SFPmaxBitRateGbps,
            'temperature': SFPtemp,
            'rx-pow': SFPrxPow,
            'tx-pow': SFPtxPow,
            'status': 'ok'}

  def _checkBERTpattern(self, byt):
    for p in self._bertPatterns.keys():
      if self._bertPatterns[p] == chr(byt):
        return p
    return "unknown ('" + chr(byt) + "')"

  def getStats(self):
    self._write("r")
    time.sleep(.01)
    b = self._read(27)

    if len(b) >= 27:
      currBitRateGbps = float(b[0]*math.pow(2,24)+b[1]*math.pow(2,16)+b[2]*math.pow(2,8)+b[3])/1e8
      currBertPattern = self._checkBERTpattern(b[4])

      currSfpRxPow = float(32768.-(b[5]*256+b[6]))/100.
      currSfpTxPow = float(32768.-(b[7]*256+b[8]))/100.
      currWaveLength = float(b[9]*math.pow(2,16)+b[10]*math.pow(2,8)+b[11]) / 100.
      currSfpTemp = float(32768.-(b[12]*256+b[13]))/100.
      currSfpStatus = b[14]
      statuStr = ""
      if currSfpStatus & 0x3F == 0:
        statuStr = "SFP not in use"
      elif currSfpStatus & 0x3F == 1:
        statuStr = "SFP no signal"
      elif currSfpStatus & 0x3F == 2:
        # statuStr = "SFP signal and synch OK"
        pass
      elif currSfpStatus & 0x3F == 3:
        statuStr = "SFP signal but no lock"
      else:
        msg = "SFP wtf status (0x%2.2X)" % currSfpStatus
        return {'status': msg}

      if currSfpStatus & 0x40 == 0:
        statuStr = "no SFP detected"
      if currSfpStatus & 0x80:
        statuStr = "new SFP inserted"

      currBitCount = float((b[15]*math.pow(2,16)+b[16]*math.pow(2,8)+b[17])*math.pow(2,b[18]-24))
      currBitErrCount = float((b[19]*math.pow(2,16)+b[20]*math.pow(2,8)+b[21])*math.pow(2,b[22]-24))
      if currBitCount <= self._prevBitCount or currBitCount == 0:
        currBER = -1.
      else:
        currBER = currBitErrCount/currBitCount
      self._prevBitCount = currBitCount

      currEyeHoriz = float(b[23])/32.
      currEyeVert = float(b[24])*3.125

      if b[26] != 0x00:
        msg = "unexpected termination (0x%2.2x != 0x00)" % b[26]
        return {'status': msg}
      
      elapsedTime = str(datetime.now()-self.timeStart)[:-3]
      d = {'datetime': str(datetime.utcnow().isoformat("T"))+"Z", # RFC3339Nano format
           'duration': elapsedTime,
           'ber': currBER,
           'bit-cnt': currBitCount,
           'err-cnt': currBitErrCount,
           'rate-gpbs': currBitRateGbps,
           'pattern': currBertPattern,
           'sfp-rx-pow': currSfpRxPow,
           'sfp-tx-pow': currSfpTxPow,
           'sfp-temp': currSfpTemp,
           'eye-horz': currEyeHoriz,
           'eye-vert': currEyeVert,
           'status': 'ok' if statuStr == '' else statuStr}
    else:
      d = {'status': "un-expected reply from tester"}
    
    return d

  def setWaveLength(self, wl_nm):
    self._write("setwl " + str(wl_nm))

  def setDataRate(self, rate_Gbps):
    self._write("setrate " + str(int(float(rate_Gbps)*1e6))) # kbps
    # this will effectively restart the test, so let's make it obvious
    self.resetStats()

  def setPattern(self, patt):
    if patt in self._bertPatterns.keys():
      self._write("setpat " + self._bertPatterns[patt])
      # this will effectively restart the test, so let's make it obvious
      self.resetStats()

      return True
    else:
      return False

  def setSFPtxEnable(self, on: bool):
    cmdV = "1" if on else "0"
    self._write("tx " + cmdV)

  def resetStats(self):
    self._prevBitCount = -1
    self.timeStart = datetime.now()

    self._write("reset")

  def runQuickTest(self):
    print("\nBERT run quick test")
    self._write("test")
    print("test running.. ")
    time.sleep(20)
    b = (self._read()).decode()
    print(b)


def __print(d, useJson):
  endWith = '\n'
  if useJson:
    strData = json.dumps(d)
  else:
    if 'model' in d.keys():
      strData = "BERT Model: %s - Version %s" % (d['model'], d['version'])
    elif 'distance-km' in d.keys():
      strData = "SFP: %s (%s, %s) - %.1fkm %.1fnm %.1fdegC Rx %.2fdBm Tx %.2fdBm" % \
        (d['serial'], d['vendor'], d['part-num'], d['distance-km'], d['wavelength'], \
        d['temperature'], d['rx-pow'], d['tx-pow'])
    elif 'ber' in d.keys():
      strData = "%12s\tBER: %13e (errCnt: %12e) eye: %4.2fUI %6.2fmV %22s" % \
        (d['duration'], d['ber'], d['err-cnt'], d['eye-horz'], d['eye-vert'], d['status'])
      endWith = '\r'
    elif 'status' in d.keys():
      strData = d['status']
    else:
      # unknown ??
      return

  print(strData, end=endWith)

if __name__ == "__main__":
  # parse arguments. serial port path needed
  parser = argparse.ArgumentParser()
  parser.add_argument('port', help="serial port")
  parser.add_argument('-r', '--rate', type=float, help="datarate in Gbps. default taken from sfp")
  parser.add_argument('-p', '--pattern', help="bert pattern as \"PRBS<7|9|11|15|23|31|58|63>\". default \"PRBS7\"")
  parser.add_argument('-f', '--frequency', type=float, default=1., help="polling frequency. default 1Hz")
  parser.add_argument('-j', '--json', action='store_true', help="output as json")

  args = parser.parse_args()
  serPort = args.port
  pollTimeout = 1./args.frequency
  useJson = args.json
  datarate = args.rate
  pattern = args.pattern

  tester = EyeBERT_MicroX(serPort, datarate=datarate, pattern=pattern)

  __print(tester.getTesterInfo(), useJson)
  __print(tester.getSFPinfo(), useJson)

  userInput = []
  while 1:
    # wait for inputs from user. if timeout print test stats
    r = select.select([sys.stdin], [], [], pollTimeout)

    if sys.stdin in r[0]:
      # read line from input and strip off newline char
      l = sys.stdin.readline()[:-1]

      # allow user to interact with the test
      if l == '':
        userInput = []
      elif l == 'q' or l == 'quit':
        tester.Close()
        exit(0)
      elif l == 'r' or l == 'reset':
        tester.resetStats()
      elif l == 'sfp':
        tester.getSFPinfo()
      elif 'tx' in l:
        txS = l.split()
        if len(txS) == 2:
          try:
            if txS[1].lower() == 'on':
              tester.setSFPtxEnable(True)
            if txS[1].lower() == 'off':
              tester.setSFPtxEnable(False)
          except:
            pass
      elif 'rate' in l: # this will reset the test
        rateS = l.split()
        if len(rateS) == 2:
          try:
            rateV = float(rateS[1])
            tester.setDataRate(rateV)
          except:
            pass
      elif 'p' in l: # this will reset the test
        pattS = l.split()
        if len(pattS) == 2:
          if pattS[1] in tester._bertPatterns.keys():
            tester.setPattern(pattS[1])
      else:
        userInput.append(l)
    else:
      __print(tester.getStats(), useJson)
