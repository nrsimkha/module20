/* Напишите код, реализующий пайплайн, работающий с целыми числами и состоящий из следующих стадий:

Стадия фильтрации отрицательных чисел (не пропускать отрицательные числа).
Стадия фильтрации чисел, не кратных 3 (не пропускать такие числа), исключая также и 0.
Стадия буферизации данных в кольцевом буфере с интерфейсом, соответствующим тому, который был дан в качестве задания в 19 модуле. В этой стадии предусмотреть опустошение буфера (и соответственно, передачу этих данных, если они есть, дальше) с определённым интервалом во времени. Значения размера буфера и этого интервала времени сделать настраиваемыми (как мы делали: через константы или глобальные переменные).
Написать источник данных для конвейера. Непосредственным источником данных должна быть консоль.

Также написать код потребителя данных конвейера. Данные от конвейера можно направить снова в консоль построчно, сопроводив их каким-нибудь поясняющим текстом, например: «Получены данные …». */

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const bufferDrainInterval = 5 * time.Second

const bufferSize = 2

var inputCounter int

type RingBuffer struct {
	data       []int
	size       int
	lastInsert int
	nextRead   int
	emitTime   time.Time
}

func newRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data:       make([]int, size),
		size:       size,
		lastInsert: -1,
		emitTime:   time.Time{},
	}
}

func (r *RingBuffer) Insert(input int) {
	r.lastInsert = (r.lastInsert + 1) % r.size
	r.data[r.lastInsert] = input

	if r.nextRead == r.lastInsert {
		r.nextRead = (r.nextRead + 1) % r.size
	}
}

func (r *RingBuffer) Emit() []int {
	output := []int{}
	for {
		if r.data[r.nextRead] != 0 {
			output = append(output, r.data[r.nextRead])
			r.data[r.nextRead] = 0
		}
		if r.nextRead == r.lastInsert || r.lastInsert == -1 {
			break
		}
		r.nextRead = (r.nextRead + 1) % r.size
	}
	return output
}

func noNegative(done <-chan bool, input <-chan int) <-chan int {
	noNegativeStream := make(chan int)
	go func() {
		defer close(noNegativeStream)
		for {
			select {
			case <-done:
				return
			case i, isChannelOpen := <-input:
				if !isChannelOpen {
					return
				}
				if i > 0 {
					select {
					case noNegativeStream <- i:
						if !isChannelOpen {
							return
						}
					case <-done:
						return
					}
				}

			}
		}

	}()
	return noNegativeStream
}

func filterDividedByThree(done <-chan bool, input <-chan int) <-chan int {
	dividedByThree := make(chan int)
	go func() {
		defer close(dividedByThree)
		for {
			select {
			case <-done:
				return
			case i, isChannelOpen := <-input:
				if !isChannelOpen {
					return
				}
				if i%3 == 0 {
					select {
					case dividedByThree <- i:
						if !isChannelOpen {
							return
						}
					case <-done:
						return
					}
				}

			}
		}

	}()
	return dividedByThree
}
func bufferedInt(done <-chan bool, input <-chan int) <-chan int {
	output := make(chan int)

	buffer := newRingBuffer(bufferSize)
	go func() {
		for {
			select {
			case data := <-input:
				buffer.Insert(data)
			case <-done:
				return
			}
		}
	}()
	go func() {
		for {
			select {
			case <-time.After(bufferDrainInterval):
				bufferData := buffer.Emit()

				if bufferData != nil {
					for _, data := range bufferData {
						select {
						case output <- data:
						case <-done:
							return
						}
					}
				}
			case <-done:
				return
			}
		}
	}()
	return output

}

func main() {

	producer := func() (<-chan int, <-chan bool) {
		c := make(chan int)
		done := make(chan bool)
		go func() {
			defer close(done)
			fmt.Println("Введите числа")
			scanner := bufio.NewScanner(os.Stdin)
			var data string
			for {
				scanner.Scan()
				data = scanner.Text()
				if strings.EqualFold(data, "exit") {
					fmt.Println("Программа завершила работу!")
					return
				}
				i, err := strconv.Atoi(data)
				if err != nil {
					fmt.Println("Программа обрабатывает только целые числа!")
					continue
				}
				c <- i
			}
		}()
		return c, done
	}
	source, done := producer()

	pipeline := bufferedInt(done, filterDividedByThree(done, noNegative(done, source)))

	consumer := func(done <-chan bool, c <-chan int) {

		for {
			select {
			case data := <-c:

				fmt.Printf("Обработаны данные: %d\n", data)
			case <-done:
				return
			}
		}

	}
	consumer(done, pipeline)
}
