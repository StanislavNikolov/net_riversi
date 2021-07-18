#include <iostream>

std::string board;
int msLeft;

void readBoard() {
	std::cin >> msLeft >> board;
}

void makeMove(int row, int column) {
	std::cout << row*8 + column << std::endl;
}

int main() {

	while(true) {
		readBoard();
		std::cerr << "Board read successfuly" << std::endl;
		makeMove(3, 5);
	}

	return 0;
}

