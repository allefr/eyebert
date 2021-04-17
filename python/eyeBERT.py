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
  getBERTinfo()
    returns dict of model and version of the tester
  getSFPinfo()
    returns dict with DDM info from the SFP
  BERTreadStats()
    returns dict with ongoing test stats
  setBERTwaveLength()
    sets wavelength in nm to use during the BER test
  setBERTdataRate()
    sets datarate in Gbps to use during the BER test
  setBERTpattern(pattern)
    sets pattern to use during the BER test. `pattern` must be one of:
    - pattLowPow
    - pattPRBS7
    - pattPRBS9
    - pattPRBS11
    - pattPRBS15
    - pattPRBS23
    - pattPRBS31
    - pattPRBS58
    - pattPRBS63
    - pattLoopBk
  setSFPtxEnable(on)
    turns SFP tx On or Off
  BERTrestartTest()
    resets all test stats, effectively re-starting the BER test
  BERTrunQuickTest()
    runs quick test using datarates ranging from what the SFP supports
  """

  _bertPatterns = {pattLowPow: "0",
                   pattPRBS7: "7",   
                   pattPRBS9: "9",  # PRBS 2^9-1
                   pattPRBS11: "1", # PRBS 2^11-1
                   pattPRBS15: "5", # PRBS 2^15-1
                   pattPRBS23: "2", # PRBS 2^23-1
                   pattPRBS31: "3", # PRBS 2^31-1
                   pattPRBS58: "8", # PRBS 2^58-1
                   pattPRBS63: "6", # PRBS 2^63-1
                   pattLoopBk: "L", # data on the input is retransmitted on the output
  }

  def __init__(self, port: str,
               wavelen=None,
               datarate=None,
               pattern=None,
               noOutput: bool=True,
               useJson: bool=False,
               ):
    """
    Parameters
    ----------
    port : str
      path to the serial port connected to the Tester
    noOutput : bool, optional
      by default do not print stuff
    useJson : bool, optional
      output as json
    """
    self.serPort = port
    self._ser = self._open()

    self._json = useJson
    self._noOutput = noOutput

    # BERT info
    self.getBERTinfo()

    if wavelen is not None:
      self.setBERTwaveLength(wavelen)
    if datarate is not None:
      self.setBERTdataRate(datarate)
    if pattern is not None:
      self.setBERTpattern(pattern)

    self.timeStart = datetime.now()
    self._prevBitCount = -1

  def _open(self):
    # note here baudrate can be anything
    ser = serial.Serial(port=self.serPort, baudrate=500000, timeout=1)
    if not ser.isOpen():
      msg = "can't open serial port! ('%s')" % self.serPort
      self.__print(msg, {'status': msg})
      exit(-1)
    ser.flushInput()
    ser.flushOutput()
    # get rid of any garbage in the serial buffer
    while ser.inWaiting():
      ser.read()
    time.sleep(.1)

    return ser

  def _close(self):
    self._ser.close()

  def _write(self, cmd_string):
    while self._ser.inWaiting():
      self._ser.read()
    # do not use "=", it crashes the tester
    self._ser.write((cmd_string.replace("=", " ") + "\r\n").encode())

  def _read(self, num=None):
    buff = bytearray([])
    while self._ser.inWaiting():
      buff.extend(self._ser.read())
    if num is not None and len(buff) < num:
      # try again after some time
      time.sleep(.5)
      while self._ser.inWaiting():
        buff.extend(self._ser.read())
    # print(len(buff))
    if len(buff) == 0:
      err = "tester seems stuck! please power cycle it"
      self.__print(err, {'status': err})
      self._close()
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

  def getBERTinfo(self):
    self._write("?")
    time.sleep(.05)
    b = (self._read()).decode()
    # print(b)
    strn = b.split()
    if len(strn) >= 3:
      BERTmodel = strn[0] + " " + strn[1]
      BERTversion = strn[2].replace(":", "")
      status = 'ok'
      msg = "model %s, version: %s" % (BERTmodel, BERTversion)
    else:
      BERTmodel = None
      BERTversion = None
      status = "un-expected reply from tester"
      msg = status
    
    d = {'model': BERTmodel, 'version': BERTversion, 'status': status}
    self.__print(msg, d)
    return d

  def getSFPinfo(self):
    self._write("sfp")
    time.sleep(.2)
    # total bytes to receive is around 2303, but only care about first part
    # still need to wait till all the data is received, or will fuck up other commands.. :(
    b = (self._read(2250)).decode()
    # print(b)
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
        if "Media" in ln:
          SFPdistanceKm = float(ln.split()[2])
        if "Bit Rate" in ln:
          SFPminBitRateGbps = float(ln.split()[2])
          SFPmaxBitRateGbps = float(ln.split()[4])
    elif "No SFP Inserted" in strn:
      msg = "no SFP detected"
      d = {'status': msg}
    else:
      msg = "un-expected reply from tester"
      d = {'status': msg}

    if msg is None:
      msg = "SFP: %s (%s, %s) - %dkm %.1fnm %.1fdegC Rx %.2fdBm Tx %.2fdBm" % (SFPserial, SFPvendor, SFPpartNum, SFPdistanceKm, SFPwaveLen, SFPtemp, SFPrxPow, SFPtxPow)
      d = {'vendor': SFPvendor,
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

    self.__print(msg, d)
    return d

  def _checkBERTpattern(self, byt):
    for p in self._bertPatterns.keys():
      if self._bertPatterns[p] == chr(byt):
        return p
    return "unknown ('" + chr(byt) + "')"

  def BERTreadStats(self):
    self._write("r")
    time.sleep(.005)
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
        d = {'status': msg}
        self.__print("\n%s" % msg, d)
        return d

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
        d = {'status': msg}
        self.__print(msg, d)
        return d
      
      elapsedTime = str(datetime.now()-self.timeStart)[:-3]
      msg = "%s\tBERT: %13e (errCnt: %12e) eye: %4.2fUI %6.2fmV %22s" % (elapsedTime, currBER, currBitErrCount, currEyeHoriz, currEyeVert, statuStr)
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
      self.__print(msg, d, False if self._json else True)
    else:
      msg = "un-expected reply from tester"
      d = {'status': msg}
      self.__print("\n%s\n" % msg, d)
    
    return d

  def setBERTwaveLength(self, wl_nm):
    msg = "setting BERT wavelength to %.2f nm.." % float(wl_nm)
    d = {'setting': {'bert-wavelength': float(wl_nm)}}
    self.__print(msg, d)
    self._write("setwl " + str(wl_nm))

  def setBERTdataRate(self, rate_Gbps):
    msg = "setting BERT datarate to %d Gbps.." % float(rate_Gbps)
    d = {'setting': {'bert-datarate': float(rate_Gbps)}}
    self.__print(msg, d)
    self._write("setrate " + str(int(float(rate_Gbps)*1e6))) # kbps
    # this will effectively restart the test, so le't make it obvious
    self.BERTrestartTest()

  def setBERTpattern(self, patt):
    if patt in self._bertPatterns.keys():
      msg = "setting BERT pattern to %s" % (patt)
      d = {'setting': {'bert-pattern': patt}}
      self.__print(msg, d)
      self._write("setpat " + self._bertPatterns[patt])
      # this will effectively restart the test, so le't make it obvious
      self.BERTrestartTest()
    else:
      msg = "invalid pattern '%s'" % patt
      self.__print(msg, {'status': msg})

  def setSFPtxEnable(self, on: bool):
    msg = "setting SFP tx enable: %s" % str(on)
    d = {'setting': {'sfp-txEnable': on}}
    self.__print(msg, d)
    cmdV = "1" if on else "0"
    self._write("tx " + cmdV)

  def BERTrestartTest(self):
    msg = "test reset"
    d = {'status': msg}
    self.__print(msg, d)
    self._write("reset")
    self._prevBitCount = -1
    self.timeStart = datetime.now()

  def BERTrunQuickTest(self):
    print("\nBERT run quick test")
    self._write("test")
    print("test running.. ")
    time.sleep(20)
    b = (self._read()).decode()
    print(b)

if __name__ == "__main__":
  # parse arguments. serial port path needed
  parser = argparse.ArgumentParser()
  parser.add_argument('port', help="serial port path")
  parser.add_argument('-r', '--rate', type=float, help="datarate in Gbps. default taken from sfp")
  parser.add_argument('-w', '--wavelen', type=float, help="wavelength in nm. default taken from sfp")
  parser.add_argument('-p', '--pattern', help="bert pattern as \"PRBS<7|9|11|15|23|31|58|63>\". default \"PRBS7\"")
  parser.add_argument('-f', '--frequency', type=float, default=1., help="polling frequency. default 1Hz")
  parser.add_argument('-s', '--silent', action='store_true', help="no printed output")
  parser.add_argument('-j', '--json', action='store_true', help="output as json")

  args = parser.parse_args()
  serPort = args.port
  pollTimeout = 1./args.frequency
  useJson = args.json
  silentPrint = args.silent
  datarate = args.rate
  wavelen = args.wavelen
  pattern = args.pattern

  tester = EyeBERT_MicroX(serPort,
                          wavelen=wavelen,
                          datarate=datarate,
                          pattern=pattern,
                          noOutput=silentPrint,
                          useJson=useJson,
                          )

  tester.getSFPinfo()
  tester.BERTrestartTest()

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
        tester._close()
        exit(0)
      elif l == 'r' or l == 'reset':
        tester.BERTrestartTest()
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
            tester.setBERTdataRate(rateV)
          except:
              pass
      elif 'p' in l: # this will reset the test
        pattS = l.split()
        if len(pattS) == 2:
          if pattS[1] in tester._bertPatterns.keys():
            tester.setBERTpattern(pattS[1])
      else:
        userInput.append(l)
    else:
      tester.BERTreadStats()
