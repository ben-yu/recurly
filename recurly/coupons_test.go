package recurly

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

// TestCouponsEncoding ensures structs are encoded to XML properly.
// Because Recurly supports partial updates, it's important that only defined
// fields are handled properly -- including types like booleans and integers which
// have zero values that we want to send.
func TestCoupons_Encoding(t *testing.T) {
	redeem, _ := time.Parse(datetimeFormat, "2014-01-01T07:00:00Z")
	tests := []struct {
		v        Coupon
		expected string
	}{
		{
			v:        Coupon{},
			expected: "<coupon><coupon_code></coupon_code><name></name><discount_type></discount_type></coupon>",
		},
		{
			v: Coupon{
				Code:         "special",
				Name:         "Special 10% off",
				DiscountType: "percent",
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><discount_type>percent</discount_type></coupon>",
		},
		{
			v: Coupon{
				Code:              "special",
				Name:              "Special 10% off",
				HostedDescription: "Save 10%",
				DiscountType:      "percent",
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><hosted_description>Save 10%</hosted_description><discount_type>percent</discount_type></coupon>",
		},
		{
			v: Coupon{
				Code:               "special",
				Name:               "Special 10% off",
				InvoiceDescription: "Coupon: Special 10% off",
				DiscountType:       "percent",
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><invoice_description>Coupon: Special 10% off</invoice_description><discount_type>percent</discount_type></coupon>",
		},
		{
			v: Coupon{
				Code:         "special",
				Name:         "Special 10% off",
				DiscountType: "percent",
				RedeemByDate: NewTime(redeem),
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><discount_type>percent</discount_type><redeem_by_date>2014-01-01T07:00:00Z</redeem_by_date></coupon>",
		},
		{
			v: Coupon{
				Code:         "special",
				Name:         "Special 10% off",
				DiscountType: "percent",
				SingleUse:    NewBool(true),
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><discount_type>percent</discount_type><single_use>true</single_use></coupon>",
		},
		{
			v: Coupon{
				Code:             "special",
				Name:             "Special 10% off",
				DiscountType:     "percent",
				AppliesForMonths: NewInt(3),
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><discount_type>percent</discount_type><applies_for_months>3</applies_for_months></coupon>",
		},
		{
			v: Coupon{
				Code:           "special",
				Name:           "Special 10% off",
				DiscountType:   "percent",
				MaxRedemptions: NewInt(20),
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><discount_type>percent</discount_type><max_redemptions>20</max_redemptions></coupon>",
		},
		{
			v: Coupon{
				Code:              "special",
				Name:              "Special 10% off",
				DiscountType:      "percent",
				AppliesToAllPlans: NewBool(false),
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><discount_type>percent</discount_type><applies_to_all_plans>false</applies_to_all_plans></coupon>",
		},
		{
			v: Coupon{
				Code:            "special",
				Name:            "Special 10% off",
				DiscountType:    "percent",
				DiscountPercent: 10,
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><discount_type>percent</discount_type><discount_percent>10</discount_percent></coupon>",
		},
		{
			v: Coupon{
				Code:            "special",
				Name:            "Special $10 off",
				DiscountType:    "dollars",
				DiscountPercent: 1000,
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special $10 off</name><discount_type>dollars</discount_type><discount_percent>1000</discount_percent></coupon>",
		},
		{
			v: Coupon{
				Code:              "special",
				Name:              "Special 10% off",
				DiscountType:      "percent",
				AppliesToAllPlans: NewBool(false),
				PlanCodes: &[]CouponPlanCode{
					CouponPlanCode{Code: "gold"},
					CouponPlanCode{Code: "silver"},
				},
			},
			expected: "<coupon><coupon_code>special</coupon_code><name>Special 10% off</name><discount_type>percent</discount_type><applies_to_all_plans>false</applies_to_all_plans><plan_codes><plan_code>gold</plan_code><plan_code>silver</plan_code></plan_codes></coupon>",
		},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		if err := xml.NewEncoder(&buf).Encode(tt.v); err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if buf.String() != tt.expected {
			t.Fatalf("unexpected coupon: %v", buf.String())
		}
	}
}

func TestCoupons_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/coupons", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(200)
		io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
        <coupons type="array">
        	<coupon href="https://your-subdomain.recurly.com/v2/coupons/special">
        		<redemptions href="https://your-subdomain.recurly.com/v2/coupons/special/redemptions"/>
        		<coupon_code>special</coupon_code>
        		<name>Special 10% off</name>
        		<state>redeemable</state>
        		<discount_type>percent</discount_type>
        		<discount_percent type="integer">10</discount_percent>
        		<redeem_by_date type="datetime">2014-01-01T07:00:00Z</redeem_by_date>
        		<single_use type="boolean">true</single_use>
        		<applies_for_months nil="nil"></applies_for_months>
        		<max_redemptions type="integer">10</max_redemptions>
        		<applies_to_all_plans type="boolean">false</applies_to_all_plans>
        		<created_at type="datetime">2011-04-10T07:00:00Z</created_at>
        		<plan_codes type="array">
        			<plan_code>gold</plan_code>
        			<plan_code>platinum</plan_code>
        		</plan_codes>
        		<a name="redeem" href="https://your-subdomain.recurly.com/v2/coupons/special/redeem" method="post"/>
        	</coupon>
        </coupons>`)
	})

	r, coupons, err := client.Coupons.List(Params{"per_page": 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if r.IsError() {
		t.Fatal("expected list coupons to return OK")
	} else if len(coupons) != 1 {
		t.Fatalf("unexpected number of coupons: %d", len(coupons))
	} else if pp := r.Request.URL.Query().Get("per_page"); pp != "1" {
		t.Fatalf("unexpected per_page: %s", pp)
	}

	ts, _ := time.Parse(datetimeFormat, "2011-04-10T07:00:00Z")
	redeem, _ := time.Parse(datetimeFormat, "2014-01-01T07:00:00Z")
	if !reflect.DeepEqual(coupons, []Coupon{
		{
			XMLName:           xml.Name{Local: "coupon"},
			Code:              "special",
			Name:              "Special 10% off",
			State:             "redeemable",
			DiscountType:      "percent",
			DiscountPercent:   10,
			RedeemByDate:      NewTime(redeem),
			SingleUse:         NewBool(true),
			MaxRedemptions:    NewInt(10),
			AppliesToAllPlans: NewBool(false),
			CreatedAt:         NewTime(ts),
			PlanCodes: &[]CouponPlanCode{
				CouponPlanCode{Code: "gold"},
				CouponPlanCode{Code: "platinum"},
			},
		},
	}) {
		t.Fatalf("unexpected coupons: %v", coupons)
	}
}

func TestCoupons_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/coupons/special", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(200)
		io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
            <coupon href="https://your-subdomain.recurly.com/v2/coupons/special">
        		<redemptions href="https://your-subdomain.recurly.com/v2/coupons/special/redemptions"/>
        		<coupon_code>special</coupon_code>
        		<name>Special 10% off</name>
        		<state>redeemable</state>
        		<discount_type>percent</discount_type>
        		<discount_percent type="integer">10</discount_percent>
        		<redeem_by_date type="datetime">2014-01-01T07:00:00Z</redeem_by_date>
        		<single_use type="boolean">true</single_use>
        		<applies_for_months nil="nil"></applies_for_months>
        		<max_redemptions type="integer">10</max_redemptions>
        		<applies_to_all_plans type="boolean">false</applies_to_all_plans>
        		<created_at type="datetime">2011-04-10T07:00:00Z</created_at>
        		<plan_codes type="array">
        			<plan_code>gold</plan_code>
        			<plan_code>platinum</plan_code>
        		</plan_codes>
        		<a name="redeem" href="https://your-subdomain.recurly.com/v2/coupons/special/redeem" method="post"/>
        	</coupon>`)
	})

	r, coupon, err := client.Coupons.Get("special")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if r.IsError() {
		t.Fatal("expected get coupon to return OK")
	}

	ts, _ := time.Parse(datetimeFormat, "2011-04-10T07:00:00Z")
	redeem, _ := time.Parse(datetimeFormat, "2014-01-01T07:00:00Z")
	if !reflect.DeepEqual(coupon, Coupon{
		XMLName:           xml.Name{Local: "coupon"},
		Code:              "special",
		Name:              "Special 10% off",
		State:             "redeemable",
		DiscountType:      "percent",
		DiscountPercent:   10,
		RedeemByDate:      NewTime(redeem),
		SingleUse:         NewBool(true),
		MaxRedemptions:    NewInt(10),
		AppliesToAllPlans: NewBool(false),
		CreatedAt:         NewTime(ts),
		PlanCodes: &[]CouponPlanCode{
			CouponPlanCode{Code: "gold"},
			CouponPlanCode{Code: "platinum"},
		},
	}) {
		t.Fatalf("unexpected coupon: %v", coupon)
	}
}

func TestCoupons_Create(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/coupons", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(201)
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><coupon></coupon>`)
	})

	r, _, err := client.Coupons.Create(Coupon{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if r.IsError() {
		t.Fatal("expected create coupon to return OK")
	}
}

func TestCoupons_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/coupons/special", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(204)
	})

	r, err := client.Coupons.Delete("special")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if r.IsError() {
		t.Fatal("expected deleted coupon to return OK")
	}
}
