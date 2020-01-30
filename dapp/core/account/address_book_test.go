package account

import (
	"testing"

	"github.com/ProxeusApp/storage-app/dapp/core/embdb"
)

func TestCreateAndUpdate(t *testing.T) {
	memoryDB := embdb.OpenDummyDB()
	addressBook := &AddressBook{
		book: map[string]*AddressBookEntry{},
		db:   memoryDB,
	}

	_, err := addressBook.Create("hans", "0x11")
	if err.Error() != "ethAddress: invalid" {
		t.Error(err)
	}

	addressBookEntry, err := addressBook.Create("iana", "0xa80899bb12e4afe9787425a5e5fe166234b88185")
	if err != nil {
		t.Error(err)
	}

	addressBookEntries, err := addressBook.List("0x00")

	if len(addressBookEntries) != 1 {
		t.Error("Expected there to be 1 addressBookEntries but got: ", len(addressBookEntries))
	}

	if addressBookEntries[0].Name != addressBookEntry.Name {
		t.Errorf("Expected first addressBookEntries element to be '%s' but got: '%s'", addressBookEntry.Name, addressBookEntries[0].Name)
	}

	addressBook.Update("jig", "0xa80899bb12e4afe9787425a5e5fe166234b88185", "")
	result := addressBook.Get("0xa80899bb12e4afe9787425a5e5fe166234b88185")
	if result.Name != "jig" {
		t.Errorf("Expected Name to be updated to 'jig' but got '%s'", result.Name)
	}
	if result.PGPPublicKey != "" {
		t.Errorf("Expected Name to be updated to '' but got '%s'", result.PGPPublicKey)
	}
}

func TestListAndGet(t *testing.T) {

	addressBookEntry1 := NewAddressBookEntry("peter", "0x00", "pubKey1")
	addressBookEntry2 := NewAddressBookEntry("hans", "0x01", "pubKey2")
	addressBookEntry3 := NewAddressBookEntry("juerg", "0xa80899bb12e4afe9787425a5e5fe166234b88185", "pubKey3")
	addressBookEntry3.Hidden = true

	book := map[string]*AddressBookEntry{
		addressBookEntry1.ETHAddress: addressBookEntry1,
		addressBookEntry2.ETHAddress: addressBookEntry2,
		addressBookEntry3.ETHAddress: addressBookEntry3,
	}

	addressBook := &AddressBook{
		book: book,
		db:   embdb.OpenDummyDB(),
	}

	addressBookEntries, err := addressBook.List("0x00")
	if err != nil {
		t.Error("Unable to list addressBookEntries, err: ", err.Error())
	}

	if len(addressBookEntries) != 2 {
		t.Error("Expected there to be 2 addressBookEntries but got: ", len(addressBookEntries))
	}

	if addressBookEntries[0].Name != "hans" {
		t.Error("Expected first addressBookEntries element to be 'hans' but got: ", addressBookEntries[0].Name)
	}

	expected := addressBookEntry3
	result := addressBook.Get("0xa80899Bb12E4AFe9787425a5E5fe166234B88185")
	if expected.Name != result.Name {
		t.Errorf("Expected to get %s but got %s", expected.Name, result.Name)
	}

	result = addressBook.Get("0xa80899Bb12E4AFe9787425a5E5fe166234B88180")
	if result.Name != "" {
		t.Errorf("Expected to get '%s' but got %s", "", result.Name)
	}
}

func TestIsInvalidETHAddr(t *testing.T) {
	addressBook := &AddressBook{}

	addr := ""
	if !addressBook.isInvalidETHAddr(addr) {
		t.Errorf("expected '%s' to be invalid eth address", addr)
	}

	addr = "0x"
	if !addressBook.isInvalidETHAddr("") {
		t.Errorf("expected '%s' to be invalid eth address", addr)
	}

	addr = "0xa80899Bb12E4AFe9787425a5E5fe166234B88185"
	if addressBook.isInvalidETHAddr(addr) {
		t.Errorf("expected '%s' to be valid eth address", addr)
	}
}
