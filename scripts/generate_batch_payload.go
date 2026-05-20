package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type payload struct {
	Items []item `json:"items"`
}

type item struct {
	UPIURI       string `json:"upi_uri"`
	MerchantName string `json:"merchant_name"`
	PayerName    string `json:"payer_name"`
}

var merchantBases = []string{
	"Simba Pvt Ltd",
	"IDFC Foods",
	"North Star Trading",
	"Urban Ledger",
	"Harbor Retail",
	"Blue Oak Services",
	"Silverline Motors",
	"Aster Supplies",
	"Greenfield Mart",
	"Skyline Studio",
}

var payerFirstNames = []string{
	"Aarav", "Aditi", "Anaya", "Arjun", "Diya", "Ishaan", "Kabir", "Kavya", "Maya", "Neel",
	"Nina", "Rahul", "Riya", "Sara", "Vihaan", "Zoya", "Asha", "Dev", "Esha", "Kabya",
}

var payerLastNames = []string{
	"Shah", "Mehta", "Patel", "Iyer", "Nair", "Rao", "Khan", "Gupta", "Sinha", "Verma",
	"Kapoor", "Bose", "Dutta", "Jain", "Malhotra", "Pillai", "Chopra", "Singh", "Das", "Arora",
}

var banks = []string{"idfcbank", "oksbi", "okhdfcbank", "okaxis", "paytm"}

func main() {
	var (
		count = flag.Int("n", 5000, "number of batch items to generate")
		out   = flag.String("out", "testdata/batch_payload_5000.json", "output JSON file path")
		seed  = flag.Int64("seed", time.Now().UnixNano(), "random seed")
	)
	flag.Parse()

	r := rand.New(rand.NewSource(*seed))
	items := make([]item, *count)
	for i := 0; i < *count; i++ {
		merchantBase := merchantBases[i%len(merchantBases)]
		merchantName := fmt.Sprintf("%s %04d", merchantBase, i+1)
		payerName := fmt.Sprintf("%s %s %04d", pick(r, payerFirstNames), pick(r, payerLastNames), i+1)
		upiURI := buildUPIURI(i+1, merchantName, payerName, pick(r, banks), r)

		items[i] = item{
			UPIURI:       upiURI,
			MerchantName: merchantName,
			PayerName:    payerName,
		}
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil && filepath.Dir(*out) != "." {
		fatal(err)
	}

	f, err := os.Create(*out)
	if err != nil {
		fatal(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload{Items: items}); err != nil {
		fatal(err)
	}

	fmt.Printf("wrote %d items to %s\n", len(items), *out)
}

func buildUPIURI(index int, merchantName, payerName, bank string, r *rand.Rand) string {
	values := url.Values{}
	values.Set("pa", fmt.Sprintf("merchant%04d@%s", index, bank))
	values.Set("pn", merchantName)
	values.Set("tn", fmt.Sprintf("invoice-%04d-%s", index, randomTag(r)))
	values.Set("am", fmt.Sprintf("%.2f", 50+float64(r.Intn(95000))/100))
	values.Set("cu", "INR")
	values.Set("tr", fmt.Sprintf("QRBATCH-%05d-%s", index, randomTag(r)))
	values.Set("mc", strconv.Itoa(5000+r.Intn(4000)))
	values.Set("tid", fmt.Sprintf("TID%05d%s", index, randomTag(r)))
	values.Set("note", fmt.Sprintf("%s -> %s", merchantName, payerName))
	return "upi://pay?" + values.Encode()
}

func pick(r *rand.Rand, values []string) string {
	return values[r.Intn(len(values))]
}

func randomTag(r *rand.Rand) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	var b strings.Builder
	b.Grow(8)
	for i := 0; i < 8; i++ {
		b.WriteByte(letters[r.Intn(len(letters))])
	}
	return b.String()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
