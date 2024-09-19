package utils

import "testing"

func TestParseToken(t *testing.T) {
	token := "bEkfuqCwIhqHZ93iSWRC5mq-AUpyoBp-yOCyyqQVJigODPDFX2zxVMLPkGg3-UxsVndgzvbUbtY9fA6sC-G3OVfg2IEr9QgaqPztt15y4VBfEGUEebv8pMWe1OjvsT29hnMampbF_OOL5dnItlO1031ejFg2ugSar5aeYRFiFrZanhIkFiy-h95HJlg6op-vwfUpguYfo-f1hbz4KUEvOGHvzMQyprHpjuiTPSdiGUsRw4d1l19DJTFOuUe45XQN92IkMOT5ladR1_8uLYjcIHBVcSqBd9sinAab2k0yjOC4DPENaLHcnE1ZUG2Dj4LPifWiiJTvpclpONyfAJnfkFX-9cD-JOTUIKahmusMR6h1nHA0ZMMFYUTZjZvK3IRflB3J-9wQkStk7___mJH-GqcMrMk5xE9zB4QkDRa00HuLAix5aCfBf5Vrbl70k8Lf"
	phone, err := ParseToken(token)
	if err != nil {
		t.Error(err)
	}
	t.Log(phone)
}

func TestRsaDecrypt(t *testing.T) {
	encrypted := "aCH513/PYicENt8MZaM6gC56kM7D7ipPfarC0BkGhDoWR9wwC7g+D/2GcOoZGjHke4pEJAkzhm7/wS/Od+QegmaTeSwhtvaVETF6296MSF7etG9p/pyRTJNtv154ZayN0ri8AfI3s+D0bHBAjZ92LXsI8gt6zzmCTMBOX06Tshs="
	phone, err := RsaDecrypt(encrypted)
	if err != nil {
		t.Error(err)
	}
	t.Log(phone)
}
