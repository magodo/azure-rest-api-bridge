package swagger

// We use following command to find all the used formats in Swagger (mgmt plane):
//
// Command:
// $ find specification/*/resource-manager -not -path "*/examples/*" -type f -name "*.json" | xargs -I %  bash -c "jq -r '.. | select (.format? != null) | .format |  select(type == \"string\")' < %"  | sort | uniq
//
// This is what it prints:
// arm-id
// base64url
// binary
// byte
// date
// date-time
// date-time-rfc1123
// decimal
// double
// duration
// email
// file
// float
// int32
// int64
// password
// time
// unixtime
// uri
// url
// uuid

type Rnd struct {
	rawString  string
	rawInteger int64
	rawNumber  float64
	rawBolean  bool

	// With format modifier
	armid           string
	base64url       string
	binary          string
	byte            string
	date            string
	dateTime        string
	dateTimeRfc1123 string
	decimal         float64
	double          float64
	duration        string
	email           string
	file            string
	float           float64
	int32           int64
	int64           int64
	password        string
	time            string
	unixtime        int64
	uri             string
	url             string
	uuid            string
}
