package main

type ASTNodeType string

const (
	TypeOperation ASTNodeType = "operation"
	TypeNumber    ASTNodeType = "number"
	TypeVariable  ASTNodeType = "variable"
)

type ASTNodeDto struct {
	Type  ASTNodeType `json:"type"` // 'operation' | 'number' | 'variable';
	Value string      `json:"value"`
	Index uint32      `json:"miner"`
	Left  *ASTNodeDto `json:"left"`
	Right *ASTNodeDto `json:"right"`
}

type ComputeRequestDto struct {
	Numbers []int      `json:"numbers"`
	Ast     ASTNodeDto `json:"ast"`
}

type ComputeResponse struct {
	Result int `json:"result"`
}

type ComputeResponseError struct {
	Error string `json:"error"`
}
