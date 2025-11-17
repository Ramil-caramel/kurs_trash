package netapi


import(
	"encoding/json"
)


type BaseRequestStruct struct{
	Command    string    `json:"command"`

}

type GetPiece struct {
    Command    string    `json:"command"`

    FileName   string    `json:"file_name"`
    PieceIndex int       `json:"piece_index"`
}


type PostPiece struct {
    Command    string    `json:"command"`

    FileName   string    `json:"file_name"`
    //PieceIndex int       `json:"piece_index"`
	Data []byte 		 `json:"data"`
}


type Error struct {
    Command    string    `json:"command"`
	Error 	   string    `json:"error"`
}


type HndShake struct {
    Command    string    `json:"command"`

    FileName   string    `json:"file_name"`
    //PieceIndex int       `json:"piece_index"`
	BitMap 	   int		 `json:"data"`
}


func ParsCommandRequest(jsonData []byte) (string, error){
	var base BaseRequestStruct
	err := json.Unmarshal(jsonData, &base)
	
    if err != nil {
		// TODO
        return "", err
    }

	return base.Command, nil

}


func CreateGetRequest(){
	var get GetPiece
	get.Command 
}
/*
func CreateReaponce (data []byte) (error){
	josn, err := json.Marshal([]data)
}
	*/