package filehandler_test

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"user/core/filehandler"
)
/*

Данные на оторых проводилось тестирование

"/home/rama/Загрузки/avidreaders.txt 00000"
"/home/rama/Загрузки/avidreaders.ru__oblomov.txt 1111111111111"
"/home.txt 1011111111111"
*/

func TestNewData(t *testing.T){
	ph := &filehandler.PublicHouse{}

	//publicHouseFile := "PublicHouse.txt"
    //os.Remove(publicHouseFile)

	testFilePath := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt"
	
	err := ph.NewData(testFilePath, 128*1024)
    if err != nil {
        t.Fatalf("NewPublicHouse returned error: %v", err)
    }

    // проверяем, что файл создан
    file, err := os.Open("PublicHouse.txt")
    if err != nil {
        t.Fatalf("cannot read PublicHouse.txt: %v", err)
    }


	scanner := bufio.NewScanner(file)
	expected := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt 1111111111111"
	found := false
    for scanner.Scan() {
        line := scanner.Text()

        if line == expected{
            found = true
        } 
    }
	if !found {
		t.Fatal("cannot find expected data in PublicHouse.txt")
	}
    if err = scanner.Err(); err != nil {
        t.Fatal("error in scaner")
    }

    //os.Remove(publicHouseFile)
}

func TestGetBitmap(t *testing.T){
	ph := &filehandler.PublicHouse{}
	testFilePath := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt"
	err := ph.NewData(testFilePath, 128*1024)
	if err != nil {
        t.Fatalf("TestGetBitmap returned error: %v", err)
    }

	tail,err := ph.GetBitmap(testFilePath)
    if err != nil {
        t.Fatalf("TestGetBitmap returned error: %v", err)
    }
	expected := "1111111111111"
	fmt.Println(tail)
	if tail != expected{
		t.Fatal("TestGetBitmap dont pass test")
	}
}

func TestHasPiece(t *testing.T){
	ph := &filehandler.PublicHouse{}
	testFilePath1 := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt"
	testFilePath2 := "/home.txt"
	err := ph.NewData(testFilePath1, 128*1024)
	if err != nil {
        t.Fatalf("TestGetBitmap returned error: %v", err)
    }
	
	f1,err := ph.HasPiece(testFilePath1,1)
	//fmt.Println(f1)
    if err != nil {
        t.Fatalf("TestHasPiece returned error: %v", err)
	}
	f2,err := ph.HasPiece(testFilePath2, 1)
    if err != nil {
        t.Fatalf("TestHasPiece returned error: %v", err)
	}
   // fmt.Println(f1,f2)
	if !f1{
		t.Fatal("TestHasPiece dont pass test1")
	}
	if f2{
		t.Fatal("TestHasPiece dont pass test2")
	}
}

func TestHasRecord(t *testing.T){
	ph := &filehandler.PublicHouse{}
	testFilePath1 := "/home/rama/Загрузки/avidreaders.ru__oblomov.txt"
	testFilePath2 := "/home111.txt"
	err := ph.NewData(testFilePath1, 128*1024)
	if err != nil {
        t.Fatalf("TestGetBitmap returned error: %v", err)
    }

	f1,err := ph.HasRecord(testFilePath1)
    if err != nil {
        t.Fatalf("TestHasRecord returned error: %v", err)
	}
	f2,err := ph.HasRecord(testFilePath2)
    if err != nil {
        t.Fatalf("TestHasRecord returned error: %v", err)
	}
    //fmt.Println(f1,f2)
	if !f1{
		t.Fatal("TestGetBitmap dont pass test1")
	}
	if f2{
		t.Fatal("TestGetBitmap dont pass test2")
	}
}