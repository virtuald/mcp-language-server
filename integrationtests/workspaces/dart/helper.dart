import 'types.dart';

/// Helper class for demonstration
class HelperClass implements SharedInterface {
  final String name;
  
  HelperClass(this.name);
  
  @override
  void doSomething() {
    print('Doing something with $name');
  }
  
  @override
  String getName() => name;
  
  /// Process method specific to HelperClass
  void process() {
    doSomething();
    print('Processing: ${getName()}');
  }
}

/// A helper function that creates instances
HelperClass createHelper(String name) {
  return HelperClass(name);
}

/// Global variable for testing
final globalHelper = HelperClass('global');