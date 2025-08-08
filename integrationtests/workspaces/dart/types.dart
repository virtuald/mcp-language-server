// Types for testing hover functionality

/// A shared class used across multiple files
/// This class demonstrates hover information
class SharedClass {
  final int value;
  
  /// Creates a new SharedClass instance
  SharedClass(this.value);
  
  /// Gets the current value
  int getValue() => value;
  
  /// A method that processes the value
  String process() {
    return 'Processed: $value';
  }
}

/// An interface for testing
abstract class SharedInterface {
  void doSomething();
  String getName();
}

/// A type alias for testing
typedef ProcessFunction = String Function(int);

/// A constant value
const String SHARED_CONSTANT = 'shared_value';

/// An enum for testing
enum Color {
  red,
  green, 
  blue
}