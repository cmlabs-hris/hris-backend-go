package validator

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"   ", true},
		{"abc", false},
		{" abc ", false},
	}
	for _, c := range cases {
		got := IsEmpty(c.input)
		if got != c.want {
			t.Errorf("IsEmpty(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestIsValidEmail(t *testing.T) {
	valid := []string{"test@example.com", "user.name+1@domain.co", "a@b.cd"}
	invalid := []string{"test@", "@example.com", "test@.com", "test@com", "test@domain", " ", ""}
	for _, email := range valid {
		if !IsValidEmail(email) {
			t.Errorf("IsValidEmail(%q) = false, want true", email)
		}
	}
	for _, email := range invalid {
		if IsValidEmail(email) {
			t.Errorf("IsValidEmail(%q) = true, want false", email)
		}
	}
}

func TestIsValidUUID(t *testing.T) {
	valid := []string{
		"0188d0f2-7b8c-7b4a-8a2b-6b8b8b8b8b8b", // valid UUIDv7
		"0188D0F2-7B8C-7B4A-8A2B-6B8B8B8B8B8B", // valid UUIDv7 (uppercase)
	}
	invalid := []string{
		"123e4567-e89b-12d3-a456-426614174000", // not v7
		"123E4567-E89B-12D3-A456-426614174000", // not v7
		"0188d0f27b8c7b4a8a2b6b8b8b8b8b8b",     // missing dashes
		"g188d0f2-7b8c-7b4a-8a2b-6b8b8b8b8b8b", // invalid hex
		"",                                     // empty
	}
	for _, uuid := range valid {
		if !IsValidUUID(uuid) {
			t.Errorf("IsValidUUID(%q) = false, want true", uuid)
		}
	}
	for _, uuid := range invalid {
		if IsValidUUID(uuid) {
			t.Errorf("IsValidUUID(%q) = true, want false", uuid)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	valid := []string{"123", "0", "9876543210"}
	invalid := []string{"abc", "123a", "", "-123"}
	for _, s := range valid {
		if !IsNumeric(s) {
			t.Errorf("IsNumeric(%q) = false, want true", s)
		}
	}
	for _, s := range invalid {
		if IsNumeric(s) {
			t.Errorf("IsNumeric(%q) = true, want false", s)
		}
	}
}

func TestIsValidDate(t *testing.T) {
	valid := []string{"2023-01-01", "2000-12-31"}
	invalid := []string{"2023-13-01", "2023-01-32", "2023/01/01", "01-01-2023", ""}
	for _, s := range valid {
		_, ok := IsValidDate(s)
		if !ok {
			t.Errorf("IsValidDate(%q) = false, want true", s)
		}
	}
	for _, s := range invalid {
		_, ok := IsValidDate(s)
		if ok {
			t.Errorf("IsValidDate(%q) = true, want false", s)
		}
	}
}

func TestIsValidNIK(t *testing.T) {
	valid := []string{"1234567890123456"}
	invalid := []string{"123456789012345", "12345678901234567", "abcdefghabcdefgh", "12345678901234ab"}
	for _, nik := range valid {
		if !IsValidNIK(nik) {
			t.Errorf("IsValidNIK(%q) = false, want true", nik)
		}
	}
	for _, nik := range invalid {
		if IsValidNIK(nik) {
			t.Errorf("IsValidNIK(%q) = true, want false", nik)
		}
	}
}

func TestIsValidPhoneNumber(t *testing.T) {
	valid := []string{"081234567890", "6281234567890", "+628123456789", "08-1234-567890", "08 1234 567890"}
	invalid := []string{"0712345678", "123456789", "0812345678901234", "abc0812345678", "0812345678a"}
	for _, phone := range valid {
		if !IsValidPhoneNumber(phone) {
			t.Errorf("IsValidPhoneNumber(%q) = false, want true", phone)
		}
	}
	for _, phone := range invalid {
		if IsValidPhoneNumber(phone) {
			t.Errorf("IsValidPhoneNumber(%q) = true, want false", phone)
		}
	}
}

func TestIsInSlice(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !IsInSlice("a", slice) {
		t.Errorf("IsInSlice('a') = false, want true")
	}
	if IsInSlice("d", slice) {
		t.Errorf("IsInSlice('d') = true, want false")
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		{Field: "email", Message: "invalid"},
		{Field: "phone", Message: "required"},
	}
	got := errs.Error()
	want := "email: invalid; phone: required"
	if got != want {
		t.Errorf("ValidationErrors.Error() = %q, want %q", got, want)
	}
}

func TestValidationErrors_ToMap(t *testing.T) {
	errs := ValidationErrors{
		{Field: "email", Message: "invalid"},
		{Field: "phone", Message: "required"},
	}
	got := errs.ToMap()
	want := map[string]string{"email": "invalid", "phone": "required"}
	if len(got) != len(want) {
		t.Errorf("ValidationErrors.ToMap() length = %d, want %d", len(got), len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("ValidationErrors.ToMap()[%q] = %q, want %q", k, got[k], v)
		}
	}
}
