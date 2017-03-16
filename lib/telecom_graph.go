package lib

type AgencyPair struct {
    Node        string
    Telecom     int
}

type TelecomGraph struct {
    Lists       []map[string] AgencyPair
}
