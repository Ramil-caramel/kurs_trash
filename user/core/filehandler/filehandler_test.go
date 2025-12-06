package filehandler_test

import (
	"bufio"
	"os"
	"testing"

	"user/core/filehandler"
)


func TestCreateFile(t *testing.T){
	filehandler.CreateFile("1234.txt",128*1024)
	file, err := os.Open("1234.txt")
    if err != nil {
        t.Fatalf("cannot read 1234.txt: %v", err)
    }
	stat,err := file.Stat()
	if err != nil {
        	t.Fatalf("TestCreateFile returned error: %v", err)
    	}
	if stat.Size() != 128*1024{
		t.Fatal("TestCreateFile dont pass test")
	}
	defer file.Close()
	os.Remove("1234.txt")
}

func TestNewData(t *testing.T){
	ph := &filehandler.PublicHouse{}

	//publicHouseFile := "PublicHouse.txt"
    //os.Remove(publicHouseFile)
	testFilePath := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}

	for _,v := range testFilePath{
		err := ph.NewSeed(v, 128*1024)
    	if err != nil {
        	t.Fatalf("NewPublicHouse returned error: %v", err)
    	}
	}

    // проверяем, что файл создан
    file, err := os.Open("PublicHouse.txt")
    if err != nil {
        t.Fatalf("cannot read PublicHouse.txt: %v", err)
    }
	defer file.Close()


	scanner := bufio.NewScanner(file)
	expected := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt|avidreaders.ru__oblomov.txt|1111111111111",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx|Рахманкулов иу8-34 лабораторная 3 отчет(2).docx|111111111111111",
	}

	found := false
    for scanner.Scan() {
        line := scanner.Text()

        if line == expected[1]{
            found = true
        } 
    }
	if !found {
		
		t.Fatal("cannot find expected data in PublicHouse.txt")
	}
	

    //os.Remove(publicHouseFile)
}

func TestGetBitmap(t *testing.T){
	ph := &filehandler.PublicHouse{}
	testFilePath := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}

	for _,v := range testFilePath{
		err := ph.NewSeed(v, 128*1024)
    	if err != nil {
        	t.Fatalf("NewPublicHouse returned error: %v", err)
    	}
	}

    // проверяем, что файл создан
    file, err := os.Open("PublicHouse.txt")
    if err != nil {
        t.Fatalf("cannot read PublicHouse.txt: %v", err)
    }
	defer file.Close()
	testData := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
		"avidreaders.ru__oblomov.txt",
		"Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}

	expected := []string{

		"1111111111111",
		"111111111111111",
		"1111111111111",
		"111111111111111",
	}
	for i,v := range testData{
		tail,err := ph.GetBitmap(v)
    	if err != nil {
        	t.Fatalf("TestGetBitmap returned error: %v", err)
    	}
		if tail != expected[i]{
			t.Fatal("TestGetBitmap dont pass test")
		}
	}
}



func TestHasPiece(t *testing.T){
	ph := &filehandler.PublicHouse{}
	testFilePath := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}

	for _,v := range testFilePath{
		err := ph.NewSeed(v, 128*1024)
    	if err != nil {
        	t.Fatalf("NewPublicHouse returned error: %v", err)
    	}
	}

	testData := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
		"avidreaders.ru__oblomov.txt",
		"avidreaders.ru__oblomov.txt",
		"Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
		"Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}
	expected := []bool{
		true,
		true,
		true,
		true,
		true,
		true,
		true,
		true,
	}
	for i,v := range testData{
		flag,_ := ph.HasPiece(v,1)
		if flag != expected[i]{
			t.Fatal("TestHasPiece dont pass test")
		}
	}

	expected = []bool{
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
	}
	for i,v := range testData{
		flag,_ := ph.HasPiece(v,20)
		if flag != expected[i]{
			t.Fatal("TestHasPiece dont pass test")
		}
	}
}

func TestHasRecord(t *testing.T){
	ph := &filehandler.PublicHouse{}
	testFilePath := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}

	for _,v := range testFilePath{
		err := ph.NewSeed(v, 128*1024)
    	if err != nil {
        	t.Fatalf("NewPublicHouse returned error: %v", err)
    	}
	}

	testData := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
		"avidreaders.ru__oblomov.txt",
		"Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
		"1234",
		"123",
		"123",
		"jsahdgrfilsfdehguoia;df/FDWSGERQAG/ASDGF",
	}

	expected := []bool{
		true,
		true,
		true,
		true,
		false,
		false,
		false,
		false,
	}
	for i,v := range testData{
		flag,_ := ph.HasRecord(v)
		if flag != expected[i]{
			t.Fatal("TestHasRecord dont pass test1")
		}
	}
}


func TestUpdateRecord(t *testing.T){
	os.Remove("PublicHouse.txt")
	ph := &filehandler.PublicHouse{}
	testFilePath := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}

	for _,v := range testFilePath{
		err := ph.NewSeed(v, 128*1024)
    	if err != nil {
        	t.Fatalf("NewPublicHouse returned error: %v", err)
    	}
	}
	testData := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}

	expected := []string{
		"000",
		"000",
	}
	for i,v := range testData{
		err := ph.UpdateRecord(v,"000")
		if err != nil {
        	t.Fatalf("TestUpdateRecord returned error: %v", err)
    	}
		bitMap,err := ph.GetBitmap(v)
		if err != nil {
        	t.Fatalf("TestUpdateRecord returned error: %v", err)
    	}
		if bitMap != expected[i]{
			t.Fatal("TestHasRecord dont pass test1")
		}
	}
}

func TestGetFullPathByName(t *testing.T){
os.Remove("PublicHouse.txt")
	ph := &filehandler.PublicHouse{}
	testFilePath := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}

	for _,v := range testFilePath{
		err := ph.NewSeed(v, 128*1024)
    	if err != nil {
        	t.Fatalf("NewPublicHouse returned error: %v", err)
    	}
	}
	testData := []string{
		"avidreaders.ru__oblomov.txt",
		"Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}

	expected := []string{
		"/home/rama/Загрузки/avidreaders.ru__oblomov.txt",
		"/home/rama/Загрузки/Рахманкулов иу8-34 лабораторная 3 отчет(2).docx",
	}
	for i,v := range testData{
		name,err := ph.GetFullPathByName(v)
		if err != nil {
        	t.Fatalf("TestGetFullPathByName returned error: %v", err)
    	}
		if name != expected[i]{
			t.Fatal("TestGetFullPathByName dont pass test1")
		}
	}
}